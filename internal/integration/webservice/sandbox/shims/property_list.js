(function (sdk) {
  'use strict';

  const _ = sdk.util;
  const PARENT = '__parent';
  const DEFAULT_MIME = 'text/plain';

  function PropertyBase (definition) {
    if (!definition || typeof definition === 'string') {
      return;
    }
    const source = (definition && definition.info) || definition;
    const meta = {};
    Object.keys(source).forEach((key) => {
      if (key.startsWith('_') && key !== '_') {
        meta[key.replace(/^_+/, '')] = source[key];
      }
    });
    if (Object.keys(meta).length) {
      this._ = _.isObject(this._) ? _.mergeDefined(this._, meta) : meta;
    }
  }

  Object.assign(PropertyBase.prototype, {
    forEachParent (options, iterator) {
      if (_.isFunction(options)) {
        iterator = options;
        options = {};
      }
      if (!_.isFunction(iterator) || !_.isObject(options)) {
        return;
      }
      let parent = this.parent();
      let grandparent = parent && _.isFunction(parent.parent) && parent.parent();
      while (parent && (grandparent || options.withRoot)) {
        iterator(parent);
        parent = grandparent;
        grandparent = grandparent && _.isFunction(grandparent.parent) && grandparent.parent();
      }
    },
    findInParents (property, customizer) {
      const owner = this.findParentContaining(property, customizer);
      return owner ? owner[property] : undefined;
    },
    findParentContaining (property, customizer) {
      let parent = this;
      do {
        if (customizer ? customizer.call(this, parent) : parent[property]) {
          return parent;
        }
        parent = parent.__parent;
      } while (parent);
      return undefined;
    },
    toJSON () {
      return _.serialise(this);
    },
    meta () {
      if (!arguments.length) {
        return this._ === undefined ? undefined : JSON.parse(JSON.stringify(this._));
      }
      const picked = {};
      Array.prototype.slice.call(arguments).forEach((key) => {
        if (this._ && key in this._) {
          picked[key] = this._[key];
        }
      });
      return picked;
    },
    parent () {
      let parent = this.__parent;
      if (parent && parent._postman_propertyIsList) {
        parent = parent.__parent || parent;
      }
      return parent || undefined;
    },
    setParent (parent) {
      _.assignHidden(this, PARENT, parent);
    }
  });

  Object.assign(PropertyBase, {
    _postman_propertyName: 'PropertyBase',
    toJSON (object) {
      return PropertyBase.prototype.toJSON.call(object);
    }
  });

  const Description = function PostmanPropertyDescription (definition) {
    if (_.isString(definition)) {
      definition = { content: definition, type: DEFAULT_MIME };
    }
    if (definition) {
      this.update(definition);
    }
  }

  Object.assign(Description.prototype, {
    update (content, type) {
      if (_.isObject(content)) {
        type = content.type;
        content = content.content;
      }
      this.content = content;
      this.type = type || DEFAULT_MIME;
    },
    toString () {
      return this.content || '';
    },
    toJSON () {
      return { content: this.content, type: this.type };
    }
  });

  Object.assign(Description, {
    _postman_propertyName: 'Description',
    isDescription (object) {
      return Boolean(object) && (object instanceof Description || _.inSuperChain(object.constructor, '_postman_propertyName', 'Description'));
    }
  });

  const Property = function PostmanProperty (definition) {
    PropertyBase.apply(this, arguments);
    const source = (definition && definition.info) || definition;
    const id = (source && source.id) || this.id || (this._ && this._.postman_id) || (this._postman_propertyRequiresId && sdk.host.uuid());
    if (id) {
      this.id = id;
    }
    if (source && source.name) {
      this.name = source.name;
    }
    if (definition && _.has(definition, 'disabled')) {
      this.disabled = Boolean(definition.disabled);
    }
    if (_.has(source, 'description')) {
      this.description = new Description(source.description);
    }
  }

  _.inherit(Property, PropertyBase);

  Object.assign(Property.prototype, {
    describe (content, type) {
      if (!Description.isDescription(this.description)) {
        this.description = new Description();
      }
      this.description.update(content, type);
    },
    findSubstitutions () {
      return Property.findSubstitutions(this.toJSON());
    }
  });

  Object.assign(Property, {
    _postman_propertyName: 'Property',
    findSubstitutions (value) {
      const found = new Set();
      const seen = new Set();
      const walk = (item) => {
        if (!item || seen.has(item)) {
          return;
        }
        if (typeof item === 'string') {
          const pattern = /\{\{([^{}]*?)}}/g;
          let match;
          while ((match = pattern.exec(item)) !== null) {
            found.add(match[0].slice(2, -2));
          }
          return;
        }
        if (typeof item === 'object') {
          seen.add(item);
          Object.keys(item).forEach((key) => walk(item[key]));
        }
      };
      walk(value);
      return Array.from(found);
    }
  });

  function unlinkFromParent (item) {
    if (item && item.__parent) {
      delete item.__parent;
    }
  }

  const PropertyList = function PostmanPropertyList (type, parent, populate) {
    PropertyBase.call(this);
    this.setParent(parent);
    Object.assign(this, { members: this.members || [], reference: this.reference || {}, Type: type });
    if (_.has(type, '_postman_propertyIndexKey')) {
      this._postman_listIndexKey = type._postman_propertyIndexKey;
    }
    if (_.has(type, '_postman_propertyIndexCaseInsensitive')) {
      this._postman_listIndexCaseInsensitive = type._postman_propertyIndexCaseInsensitive;
    }
    if (_.has(type, '_postman_propertyAllowsMultipleValues')) {
      this._postman_listAllowsMultipleValues = type._postman_propertyAllowsMultipleValues;
    }
    if (populate) {
      this.populate(populate);
    }
  }

  _.inherit(PropertyList, PropertyBase);

  Object.assign(PropertyList.prototype, {
    _postman_propertyIsList: true,
    _postman_listIndexKey: 'id',
    _postman_listIndexCaseInsensitive: false,
    _postman_listAllowsMultipleValues: false,

    indexKeyOf (item) {
      let index = item[this._postman_listIndexKey];
      if (!index) {
        return undefined;
      }
      index = String(index);
      return this._postman_listIndexCaseInsensitive ? index.toLowerCase() : index;
    },
    insert (item, before) {
      if (!_.isObject(item)) {
        return;
      }
      const duplicate = this.indexOf(item);
      if (PropertyList.isPropertyList(item[PARENT]) && item[PARENT] !== this) {
        item[PARENT].remove(item);
      }
      _.assignHidden(item, PARENT, this);
      if (duplicate > -1) {
        this.members.splice(duplicate, 1);
      }
      if (before) {
        before = this.indexOf(before);
      }
      if (before > -1) {
        this.members.splice(before, 0, item);
      }
      else {
        this.members.push(item);
      }
      const index = this.indexKeyOf(item);
      if (index === undefined) {
        return;
      }
      if (this._postman_listAllowsMultipleValues && _.has(this.reference, index)) {
        if (!Array.isArray(this.reference[index])) {
          this.reference[index] = [this.reference[index]];
        }
        this.reference[index].push(item);
        return;
      }
      this.reference[index] = item;
    },
    insertAfter (item, after) {
      return this.insert(item, this.idx(this.indexOf(after) + 1));
    },
    append (item) {
      return this.insert(item);
    },
    prepend (item) {
      return this.insert(item, this.idx(0));
    },
    add (item) {
      if (_.isNil(item) || (typeof item === 'number' && isNaN(item))) {
        return;
      }
      let created;
      if (item.constructor === this.Type) {
        created = item;
      }
      else if (_.has(this.Type, 'create')) {
        created = this.Type.create.apply(this.Type, arguments);
      }
      else {
        created = new this.Type(item);
      }
      this.insert(created);
    },
    upsert (item) {
      if (_.isNil(item) || (typeof item === 'number' && isNaN(item))) {
        return null;
      }
      const existing = this.one(item[this._postman_listIndexKey]);
      if (existing) {
        if (!_.isFunction(existing.update)) {
          throw new Error('collection: unable to upsert into a list of Type that does not support .update()');
        }
        existing.update(item);
        return false;
      }
      this.add(item);
      return true;
    },
    remove (predicate, context) {
      context = context || this;
      if (_.isString(predicate)) {
        const match = this._postman_listIndexCaseInsensitive ? predicate.toLowerCase() : predicate;
        predicate = (item) => {
          let id = item[this._postman_listIndexKey];
          if (this._postman_listIndexCaseInsensitive) {
            id = id.toLowerCase();
          }
          return id === match;
        };
      }
      else if (predicate instanceof this.Type) {
        const match = predicate;
        predicate = (item) => item === match;
      }
      if (!_.isFunction(predicate)) {
        return;
      }
      const kept = [];
      const members = this.members;
      for (let position = 0; position < members.length; position++) {
        const item = members[position];
        if (!predicate.call(context, item, position, members)) {
          kept.push(item);
          continue;
        }
        const index = this.indexKeyOf(item);
        if (index !== undefined) {
          if (this._postman_listAllowsMultipleValues && Array.isArray(this.reference[index])) {
            this.reference[index] = this.reference[index].filter((each) => each !== item);
            if (this.reference[index].length === 0) {
              delete this.reference[index];
            }
            else if (this.reference[index].length === 1) {
              this.reference[index] = this.reference[index][0];
            }
          }
          else {
            delete this.reference[index];
          }
        }
        delete item[PARENT];
      }
      members.length = 0;
      kept.forEach((item) => members.push(item));
    },
    clear () {
      this.all().forEach(unlinkFromParent);
      this.members.length = 0;
      Object.keys(this.reference).forEach((key) => { delete this.reference[key]; });
    },
    populate (items) {
      if (_.isString(items) && _.isFunction(this.Type.parse)) {
        items = this.Type.parse(items);
      }
      let list;
      if (Array.isArray(items)) {
        list = items;
      }
      else if (_.isPlainObject(items) && _.has(this.Type, 'create')) {
        Object.keys(items).forEach((key) => this.add(items[key], key, items));
        return;
      }
      else {
        list = [items];
      }
      list.forEach((item, index, all) => this.add(item, index, all));
    },
    repopulate (items) {
      this.clear();
      this.populate(items);
    },
    assimilate (source, prune) {
      const members = PropertyList.isPropertyList(source) ? source.members : source;
      const indexer = this._postman_listIndexKey;
      const sourceKeys = {};
      if (!Array.isArray(members)) {
        return;
      }
      members.forEach((item) => {
        if (!(item && _.has(item, indexer))) {
          return;
        }
        this.upsert(item);
        sourceKeys[item[indexer]] = true;
      });
      if (prune) {
        Object.keys(this.reference).forEach((key) => {
          if (!_.has(sourceKeys, key)) {
            this.remove(key);
          }
        });
      }
    },
    all () {
      return this.members.slice();
    },
    one (id) {
      const value = this.reference[this._postman_listIndexCaseInsensitive ? String(id).toLowerCase() : id];
      if (this._postman_listAllowsMultipleValues && Array.isArray(value)) {
        return value.length ? value[value.length - 1] : undefined;
      }
      return value;
    },
    get (key) {
      const member = this.one(key);
      if (!member) {
        return undefined;
      }
      return member.valueOf();
    },
    each (iterator, context) {
      if (!_.isFunction(iterator)) {
        return;
      }
      const bound = iterator.bind(context || this.__parent);
      const members = this.members.slice();
      for (let index = 0; index < members.length; index++) {
        if (bound(members[index], index, members) === false) {
          return;
        }
      }
    },
    filter (rule, context) {
      return this.members.filter(PropertyList.iteratee(rule, context));
    },
    find (rule, context) {
      return this.members.find(PropertyList.iteratee(rule, context));
    },
    map (iterator, context) {
      return this.members.map(_.isFunction(iterator) ? iterator.bind(context || this) : PropertyList.iteratee(iterator));
    },
    reduce (iterator, accumulator, context) {
      const bound = iterator.bind(context || this);
      const members = this.members;
      let index = 0;
      let total = accumulator;
      if (arguments.length < 2 && members.length) {
        total = members[0];
        index = 1;
      }
      for (; index < members.length; index++) {
        total = bound(total, members[index], index, members);
      }
      return total;
    },
    count () {
      return this.members.length;
    },
    idx (index) {
      return this.members[index];
    },
    indexOf (item) {
      return this.members.indexOf(_.isString(item) ? this.one(item) : item);
    },
    has (item, value) {
      let match = _.isString(item) ?
        this.reference[this._postman_listIndexCaseInsensitive ? item.toLowerCase() : item] :
        this.filter((member) => member === item);
      if (!match) {
        return false;
      }
      if (arguments.length === 1) {
        return Boolean(Array.isArray(match) ? match.length : match);
      }
      if (this._postman_listAllowsMultipleValues && Array.isArray(match)) {
        return match.some((each) => (_.isFunction(each.valueOf) ? each.valueOf() : each) === value);
      }
      if (_.isFunction(match.valueOf)) {
        match = match.valueOf();
      }
      return match === value;
    },
    eachParent (iterator, context) {
      if (!_.isFunction(iterator)) {
        return;
      }
      let parent = this.__parent;
      let previous;
      while (parent) {
        iterator.call(context || this, parent, previous);
        previous = parent;
        parent = parent.__parent;
      }
    },
    toObject (excludeDisabled, caseSensitive, multiValue, sanitizeKeys) {
      const object = {};
      const key = this._postman_listIndexKey;
      const sanitise = this._postman_sanitizeKeys || sanitizeKeys;
      const sensitive = !this._postman_listIndexCaseInsensitive || caseSensitive;
      const multiple = this._postman_listAllowsMultipleValues || multiValue;
      this.each((member) => {
        if (!member || !_.has(member, key) || (excludeDisabled && member.disabled) || (sanitise && !member[key])) {
          return;
        }
        const property = sensitive ? member[key] : String(member[key]).toLowerCase();
        if (multiple && _.has(object, property)) {
          if (!Array.isArray(object[property])) {
            object[property] = [object[property]];
          }
          object[property].push(member.valueOf());
          return;
        }
        object[property] = member.valueOf();
      });
      return object;
    },
    toString () {
      if (this.Type.unparse) {
        return this.Type.unparse(this.members);
      }
      return Object.prototype.toString.call(this);
    },
    toJSON () {
      if (!this.count()) {
        return [];
      }
      return this.members.map((member) => {
        if (!_.isEmpty(member) && _.isFunction(member.toJSON)) {
          return member.toJSON();
        }
        return _.serialise(member);
      });
    }
  });

  Object.assign(PropertyList, {
    _postman_propertyName: 'PropertyList',
    iteratee (rule, context) {
      if (_.isFunction(rule)) {
        return _.isObject(context) ? rule.bind(context) : rule;
      }
      if (_.isString(rule)) {
        return (member) => member !== null && member !== undefined && member[rule];
      }
      if (Array.isArray(rule)) {
        return (member) => member !== null && member !== undefined && _.isEqual(member[rule[0]], rule[1]);
      }
      if (_.isObject(rule)) {
        return (member) => Object.keys(rule).every((key) => member !== null && member !== undefined && _.isEqual(member[key], rule[key]));
      }
      return (member) => member;
    },
    isPropertyList (object) {
      return Boolean(object) && (object instanceof PropertyList || _.inSuperChain(object.constructor, '_postman_propertyName', 'PropertyList'));
    }
  });

  return { PropertyBase, Property, PropertyList, Description };
})
