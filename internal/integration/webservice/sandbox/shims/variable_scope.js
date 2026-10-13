(function (sdk) {
  'use strict';

  const _ = sdk.util;
  const { Property, PropertyBase, PropertyList } = sdk;
  const VARIABLES = /\{\{([^{}]*?)}}/g;
  const SUBREPLACE_LIMIT = 19;
  const NATIVE_TYPES = { string: true, number: true, boolean: true };

  function Substitutor (variables, defaults) {
    if (defaults) {
      variables.push(defaults);
    }
    this.variables = variables;
  }

  Object.assign(Substitutor.prototype, {
    find (key) {
      for (const source of this.variables) {
        if (!(source && _.isObject(source))) {
          continue;
        }
        if (source.constructor && source.constructor._postman_propertyName === 'VariableList') {
          const variable = source.oneNormalizedVariable(key);
          if (variable && !variable.disabled) {
            return variable;
          }
        }
        else if (_.has(source, key)) {
          return source[key];
        }
      }
      return undefined;
    },
    parse (value) {
      let text = _.isString(value) ? value : (value && _.isFunction(value.toString) && value.toString()) || '';
      let substitutions = 0;
      let replacements;
      do {
        replacements = 0;
        text = text.replace(VARIABLES, (match, token) => {
          replacements++;
          let resolved = this.find(token);
          if (resolved && _.isFunction(resolved)) {
            resolved = resolved();
          }
          if (resolved && _.isFunction(resolved.toString)) {
            resolved = resolved.toString();
          }
          return NATIVE_TYPES[typeof resolved] ? resolved : match;
        });
        if (replacements) {
          substitutions++;
        }
      } while (replacements && substitutions < SUBREPLACE_LIMIT);
      return text;
    }
  });

  Object.assign(Substitutor, {
    REGEX_EXTRACT_VARS: VARIABLES,
    VARS_SUBREPLACE_LIMIT: SUBREPLACE_LIMIT,
    DEFAULT_VARS: {},
    box (variables, defaults) {
      return variables instanceof Substitutor ? variables : new Substitutor(variables, defaults);
    },
    isInstance (subject) {
      return subject instanceof Substitutor;
    }
  });

  Substitutor.DEFAULT_VARS = new Proxy({}, {
    has (target, name) {
      return typeof name === 'string' && name.startsWith('$') && sdk.host.dynamic(name) !== undefined;
    },
    get (target, name) {
      if (typeof name !== 'string' || !name.startsWith('$')) {
        return undefined;
      }
      return () => sdk.host.dynamic(name);
    },
    getOwnPropertyDescriptor (target, name) {
      if (typeof name !== 'string' || !name.startsWith('$')) {
        return undefined;
      }
      const value = sdk.host.dynamic(name);
      return value === undefined ? undefined : { value: () => value, enumerable: true, configurable: true, writable: false };
    }
  });

  function replaceSubstitutions (text, variables) {
    if (!(text && _.isString(text))) {
      return text;
    }
    return Substitutor.box(variables, Substitutor.DEFAULT_VARS).parse(text);
  }

  function replaceSubstitutionsIn (object, variables) {
    if (!(object && _.isObject(object))) {
      return object;
    }
    const substitutor = Substitutor.box(variables, Substitutor.DEFAULT_VARS);
    const walk = (value) => {
      if (_.isString(value)) {
        return replaceSubstitutions(value, substitutor);
      }
      if (Array.isArray(value)) {
        return value.map(walk);
      }
      if (_.isPlainObject(value)) {
        const copy = {};
        Object.keys(value).forEach((key) => { copy[key] = walk(value[key]); });
        return copy;
      }
      return value;
    };
    return walk(object);
  }

  Property.replaceSubstitutions = function (text, variables) {
    if (!(text && _.isString(text))) {
      return text;
    }
    if (!Substitutor.isInstance(variables) && !Array.isArray(variables)) {
      variables = Array.prototype.slice.call(arguments, 1);
    }
    return replaceSubstitutions(text, variables);
  };
  Property.replaceSubstitutionsIn = replaceSubstitutionsIn;

  const castTo = {
    string: String,
    boolean: Boolean,
    number: Number,
    array: {
      in (value) {
        try {
          return typeof value === 'string' ? value : JSON.stringify(value);
        }
        catch (error) {
          return 'null';
        }
      },
      out (value) {
        let parsed;
        try {
          parsed = JSON.parse(value);
        }
        catch (error) {
          parsed = undefined;
        }
        return Array.isArray(parsed) ? parsed : undefined;
      }
    },
    object: {
      in (value) {
        try {
          return typeof value === 'string' ? value : JSON.stringify(value);
        }
        catch (error) {
          return 'null';
        }
      },
      out (value) {
        let parsed;
        try {
          parsed = JSON.parse(value);
        }
        catch (error) {
          parsed = undefined;
        }
        return parsed instanceof Object && !Array.isArray(parsed) ? parsed : undefined;
      }
    },
    any: {
      in (value) {
        return value;
      },
      out (value) {
        return value;
      }
    }
  };

  const Variable = function PostmanVariable (definition) {
    Property.apply(this, arguments);
    Object.assign(this, { type: 'any', value: undefined, source: undefined });
    if (!_.isNil(definition)) {
      if (_.has(definition, 'key')) {
        this.key = definition.key;
      }
      this.update(definition);
    }
  }

  _.inherit(Variable, Property);

  Object.assign(Variable.prototype, {
    get () {
      return _.isFunction(this.value) ? this.castOut(this.value()) : this.castOut(this.value);
    },
    set (value) {
      this.value = _.isFunction(value) ? value : this.castIn(value);
    },
    valueOf (value) {
      if (arguments.length) {
        this.set(value);
      }
      return this.get();
    },
    toString () {
      const value = this.valueOf();
      if (value === null) {
        return 'null';
      }
      return !_.isNil(value) && _.isFunction(value.toString) ? value.toString() : '';
    },
    cast (value) {
      return this.castOut(value);
    },
    castIn (value) {
      const handler = castTo[this.type] || castTo.any;
      return _.isFunction(handler) ? handler(value) : handler.in(value);
    },
    castOut (value) {
      const handler = castTo[this.type] || castTo.any;
      return _.isFunction(handler) ? handler(value) : handler.out(value);
    },
    valueType (typeName, noCast) {
      if (!_.isNil(typeName)) {
        typeName = typeName.toString().toLowerCase();
      }
      if (!castTo[typeName]) {
        return this.type || 'any';
      }
      this.type = typeName;
      if (!(noCast || _.isFunction(this.value))) {
        this.set(this.get());
      }
      return this.type;
    },
    update (options) {
      if (!_.isObject(options)) {
        return;
      }
      if (_.has(options, 'type')) {
        this.valueType(options.type, _.has(options, 'value'));
      }
      if (_.has(options, 'value')) {
        this.set(options.value);
      }
      if (_.has(options, 'system')) {
        this.system = options.system;
      }
      if (_.has(options, 'secret')) {
        this.secret = options.secret;
      }
      if (_.has(options, 'disabled')) {
        this.disabled = options.disabled;
      }
      if (_.has(options, 'description')) {
        this.describe(options.description);
      }
      if (_.has(options, 'source')) {
        this.source = options.source;
      }
    }
  });

  Object.assign(Variable, {
    _postman_propertyName: 'Variable',
    _postman_propertyIndexKey: 'key',
    types: castTo,
    isVariable (object) {
      return Boolean(object) && (object instanceof Variable || _.inSuperChain(object.constructor, '_postman_propertyName', 'Variable'));
    }
  });

  const VariableList = function PostmanVariableList (parent, populate) {
    PropertyList.call(this, Variable, parent, populate);
  }

  _.inherit(VariableList, PropertyList);

  Object.assign(VariableList.prototype, {
    replace (text, overrides) {
      return Property.replaceSubstitutions(text, this, overrides);
    },
    substitute (object, overrides) {
      const queue = [];
      let source = { variables: this, __parent: this.__parent };
      do {
        if (source.variables) {
          queue.push(source.variables);
        }
        source = source.__parent;
      } while (source);
      return Property.replaceSubstitutionsIn(object, queue.concat(overrides || []));
    },
    syncFromObject (object, track, prune) {
      const operations = track && { created: [], updated: [], deleted: [] };
      if (!_.isObject(object)) {
        return operations;
      }
      Object.keys(object).forEach((key) => {
        const value = object[key];
        if (this.has(key)) {
          this.one(key).set(value);
          if (operations) {
            operations.updated.push(key);
          }
          return;
        }
        const created = { value };
        created[this._postman_listIndexKey] = key;
        this.add(created);
        if (operations) {
          operations.created.push(key);
        }
      });
      if (prune !== false) {
        Object.keys(this.reference).forEach((key) => {
          if (_.has(object, key)) {
            return;
          }
          this.remove(key);
          if (operations) {
            operations.deleted.push(key);
          }
        });
      }
      return operations;
    },
    syncToObject (object) {
      if (!_.isObject(object)) {
        object = {};
      }
      Object.keys(object).forEach((key) => {
        if (!_.has(this.reference, key)) {
          delete object[key];
        }
      });
      this.each((variable) => { object[variable.key] = variable.valueOf(); });
      return object;
    },
    oneNormalizedVariable (name) {
      let variable = this.reference[name];
      if (variable && !variable.disabled) {
        return variable;
      }
      for (let index = this.members.length - 1; index >= 0; index--) {
        variable = this.members[index];
        if (variable.key === name && !variable.disabled) {
          this.reference[name] = variable;
          break;
        }
      }
      return this.reference[name];
    }
  });

  Object.assign(VariableList, {
    _postman_propertyName: 'VariableList',
    isVariableList (object) {
      return Boolean(object) && (object instanceof VariableList || _.inSuperChain(object.constructor, '_postman_propertyName', 'VariableList'));
    }
  });

  const VariableScope = function PostmanVariableScope (definition, layers) {
    if (Array.isArray(definition) || VariableList.isVariableList(definition)) {
      definition = { values: definition };
    }
    if (layers && !Array.isArray(layers)) {
      layers = [layers];
    }
    Property.call(this, definition);
    const values = definition && definition.values;
    this.values = new VariableList(this, VariableList.isVariableList(values) ? values.toJSON() : values);
    if (layers) {
      this._layers = layers.filter((layer) => VariableList.isVariableList(layer));
    }
    if (definition && definition.prefix) {
      this.prefix = definition.prefix;
    }
  }

  _.inherit(VariableScope, Property);

  Object.assign(VariableScope.prototype, {
    _postman_propertyRequiresId: true,
    variables () {
      throw new Error('`VariableScope#variables` has been discontinued, use `VariableScope#syncVariablesTo` instead.');
    },
    toObject (excludeDisabled, caseSensitive) {
      if (!this._layers) {
        return this.values.toObject(excludeDisabled, caseSensitive);
      }
      const merged = {};
      this._layers.slice().reverse().forEach((layer) => Object.assign(merged, layer.toObject(excludeDisabled, caseSensitive)));
      return Object.assign(merged, this.values.toObject(excludeDisabled, caseSensitive));
    },
    lookup (key) {
      const name = (this.prefix || '') + key;
      let variable = this.values.oneNormalizedVariable(name);
      if ((!variable || variable.disabled === true) && this._layers) {
        for (const layer of this._layers) {
          variable = layer.oneNormalizedVariable(name);
          if (variable && variable.disabled !== true) {
            break;
          }
        }
      }
      return variable && variable.disabled !== true ? variable : undefined;
    },
    has (key) {
      return Boolean(this.lookup(key));
    },
    get (key) {
      const variable = this.lookup(key);
      return variable ? variable.valueOf() : undefined;
    },
    set (key, value, options) {
      const name = (this.prefix || '') + key;
      const variable = this.values.oneNormalizedVariable(name);
      const update = { key: name, value };
      if (_.isString(options)) {
        options = { type: options };
      }
      if (options && _.isString(options.type)) {
        update.type = options.type;
      }
      if (options && typeof options.secret === 'boolean') {
        update.secret = options.secret;
      }
      if (variable && !variable.disabled) {
        variable.update(update);
        return;
      }
      this.values.add(update);
    },
    unset (key) {
      const name = (this.prefix || '') + key;
      let lastDisabled;
      this.values.remove((variable) => {
        if (variable.key !== name) {
          return false;
        }
        if (variable.disabled) {
          lastDisabled = variable;
          return false;
        }
        return true;
      });
      if (lastDisabled) {
        this.values.reference[name] = lastDisabled;
      }
    },
    clear () {
      this.values.clear();
    },
    replaceIn (template) {
      const sources = [this.values].concat(this._layers || []);
      if (_.isString(template) || Array.isArray(template)) {
        return Property.replaceSubstitutionsIn({ template }, sources).template;
      }
      if (_.isObject(template)) {
        return Property.replaceSubstitutionsIn(template, sources);
      }
      return template;
    },
    enableTracking () {
      return undefined;
    },
    disableTracking () {
      return undefined;
    },
    applyMutation (instruction, key, value, metadata) {
      if (this[instruction]) {
        this[instruction](key, value, metadata);
      }
    },
    syncVariablesFrom (object, track) {
      return this.values.syncFromObject(object, track);
    },
    syncVariablesTo (object) {
      return this.values.syncToObject(object);
    },
    toJSON () {
      const json = PropertyBase.toJSON(this);
      if (json.value) {
        json.values = json.value;
        delete json.value;
      }
      delete json._layers;
      return json;
    },
    addLayer (list) {
      if (!VariableList.isVariableList(list)) {
        return;
      }
      if (!this._layers) {
        this._layers = [];
      }
      this._layers.push(list);
    }
  });

  Object.assign(VariableScope, {
    _postman_propertyName: 'VariableScope',
    isVariableScope (object) {
      return Boolean(object) && (object instanceof VariableScope || _.inSuperChain(object.constructor, '_postman_propertyName', 'VariableScope'));
    }
  });

  return { Substitutor, Variable, VariableList, VariableScope };
})
