(function () {
  'use strict';

  class Property {
    constructor(options, separator) {
      const source = options || {};
      this.key = source.key === undefined || source.key === null ? '' : String(source.key);
      this.value = source.value === undefined ? '' : source.value;
      if (source.disabled) {
        this.disabled = true;
      }
      Object.defineProperty(this, 'separator', { value: separator, enumerable: false, writable: true });
    }

    update(options) {
      const source = options || {};
      if (source.key !== undefined) {
        this.key = String(source.key);
      }
      if (source.value !== undefined) {
        this.value = source.value;
      }
      if (source.disabled !== undefined) {
        this.disabled = Boolean(source.disabled);
      }
    }

    valueOf() {
      return this.value;
    }

    toString() {
      return this.value === null ? this.key : this.key + this.separator + this.value;
    }

    toJSON() {
      const json = { key: this.key, value: this.value };
      if (this.disabled) {
        json.disabled = true;
      }
      return json;
    }
  }

  class PropertyList {
    constructor(separator, items, caseInsensitive) {
      Object.defineProperty(this, 'members', { value: [], enumerable: false, writable: true });
      Object.defineProperty(this, 'separator', { value: separator, enumerable: false });
      Object.defineProperty(this, 'caseInsensitive', { value: Boolean(caseInsensitive), enumerable: false });
      this.populate(items);
    }

    normalize(key) {
      const text = key === undefined || key === null ? '' : String(key);
      return this.caseInsensitive ? text.toLowerCase() : text;
    }

    create(item) {
      if (item instanceof Property) {
        return item;
      }
      if (typeof item === 'string') {
        const index = item.indexOf(this.separator.trim() || this.separator);
        return index < 0
          ? new Property({ key: item, value: null }, this.separator)
          : new Property({ key: item.slice(0, index).trim(), value: item.slice(index + 1).trim() }, this.separator);
      }
      return new Property(item, this.separator);
    }

    add(item) {
      if (item === undefined || item === null) {
        return;
      }
      this.members.push(this.create(item));
    }

    append(item) {
      this.add(item);
    }

    prepend(item) {
      if (item !== undefined && item !== null) {
        this.members.unshift(this.create(item));
      }
    }

    insert(item, before) {
      const index = this.indexOf(before);
      if (index < 0) {
        this.add(item);
        return;
      }
      this.members.splice(index, 0, this.create(item));
    }

    upsert(item) {
      if (item === undefined || item === null) {
        return null;
      }
      const created = this.create(item);
      const existing = this.one(created.key);
      if (existing) {
        existing.update(item);
        return false;
      }
      this.members.push(created);
      return true;
    }

    remove(predicate, context) {
      let matches;
      if (typeof predicate === 'function') {
        matches = (member) => predicate.call(context, member);
      } else if (predicate instanceof Property) {
        matches = (member) => member === predicate;
      } else {
        const key = this.normalize(predicate);
        matches = (member) => this.normalize(member.key) === key;
      }
      this.members = this.members.filter((member) => !matches(member));
    }

    clear() {
      this.members = [];
    }

    populate(items) {
      const list = Array.isArray(items) ? items : items === undefined || items === null ? [] : [items];
      list.forEach((item) => this.add(item));
    }

    repopulate(items) {
      this.clear();
      this.populate(items);
    }

    one(key) {
      const normalized = this.normalize(key);
      for (let index = this.members.length - 1; index >= 0; index--) {
        if (this.normalize(this.members[index].key) === normalized) {
          return this.members[index];
        }
      }
      return undefined;
    }

    get(key) {
      const member = this.one(key);
      return member === undefined ? undefined : member.valueOf();
    }

    has(key, value) {
      const normalized = this.normalize(key);
      const anyValue = arguments.length < 2;
      return this.members.some((member) => this.normalize(member.key) === normalized && (anyValue || member.value === value));
    }

    idx(index) {
      return this.members[index];
    }

    indexOf(item) {
      return this.members.indexOf(item);
    }

    all() {
      return this.members.slice();
    }

    count() {
      return this.members.length;
    }

    each(iterator, context) {
      this.members.forEach(iterator, context);
    }

    map(iterator, context) {
      return this.members.map(iterator, context);
    }

    filter(iterator, context) {
      return this.members.filter(iterator, context);
    }

    find(iterator, context) {
      return this.members.find(iterator, context);
    }

    reduce(iterator, accumulator, context) {
      return this.members.reduce((total, member) => iterator.call(context, total, member), accumulator);
    }

    toObject(excludeDisabled) {
      const object = {};
      this.members.forEach((member) => {
        if (!(excludeDisabled && member.disabled)) {
          object[member.key] = member.valueOf();
        }
      });
      return object;
    }

    toString() {
      return this.members.filter((member) => !member.disabled).map((member) => member.toString()).join(this.separator === ': ' ? '\n' : '&');
    }

    toJSON() {
      return this.members.map((member) => member.toJSON());
    }
  }

  return { Property, PropertyList };
})
