(function (host, global) {
  'use strict';

  const FORMATTERS = ['NumberFormat', 'DateTimeFormat', 'PluralRules', 'Locale'];
  const DATE_DEFAULTS = { year: 'numeric', month: 'numeric', day: 'numeric' };
  const TIME_DEFAULTS = { hour: 'numeric', minute: 'numeric', second: 'numeric' };
  const COMPONENTS = ['weekday', 'era', 'year', 'month', 'day', 'dayPeriod', 'hour', 'minute', 'second', 'fractionalSecondDigits', 'timeZoneName', 'dateStyle', 'timeStyle'];
  let installed;
  let warned = false;

  function supported (locale) {
    return /^en(-|$)/i.test(String(locale));
  }

  function checked (locales) {
    if (locales === undefined || locales === null) {
      return locales;
    }
    const list = Array.isArray(locales) ? locales : [locales];
    if (list.some(supported)) {
      return locales;
    }
    if (!warned) {
      warned = true;
      host.notice('warn', 'Intl: the locale ' + JSON.stringify(list.map(String).join(', ')) + ' is not available in Xenon; en-US is used instead');
    }
    return 'en-US';
  }

  function guarded (Formatter) {
    return new Proxy(Formatter, {
      construct (target, args) {
        args[0] = checked(args[0]);
        return Reflect.construct(target, args);
      },
      apply (target, thisArg, args) {
        args[0] = checked(args[0]);
        return Reflect.apply(target, thisArg, args);
      }
    });
  }

  function load () {
    if (installed) {
      return installed;
    }
    delete global.Intl;
    const exported = host.load('intl');
    const intl = global.Intl || exported.Intl;
    if (intl.DateTimeFormat && typeof intl.DateTimeFormat.__setDefaultTimeZone === 'function') {
      intl.DateTimeFormat.__setDefaultTimeZone(host.timeZone());
    }
    FORMATTERS.forEach((name) => {
      if (intl[name] && name !== 'Locale') {
        intl[name] = guarded(intl[name]);
      }
    });
    installed = intl;
    Object.defineProperty(global, 'Intl', { value: intl, writable: true, configurable: true, enumerable: false });
    return intl;
  }

  function withDefaults (options, defaults) {
    const settings = Object.assign({}, options || {});
    if (!COMPONENTS.some((name) => settings[name] !== undefined)) {
      Object.assign(settings, defaults);
    }
    return settings;
  }

  Object.defineProperty(global, 'Intl', {
    get () {
      return load();
    },
    set (value) {
      Object.defineProperty(global, 'Intl', { value, writable: true, configurable: true, enumerable: false });
    },
    configurable: true,
    enumerable: false
  });

  function define (prototype, name, implementation) {
    Object.defineProperty(prototype, name, { value: implementation, writable: true, configurable: true, enumerable: false });
  }

  define(Number.prototype, 'toLocaleString', function toLocaleString (locales, options) {
    return new (load().NumberFormat)(locales, options).format(Number(this));
  });
  define(Date.prototype, 'toLocaleString', function toLocaleString (locales, options) {
    if (isNaN(this.getTime())) {
      return 'Invalid Date';
    }
    return new (load().DateTimeFormat)(locales, withDefaults(options, Object.assign({}, DATE_DEFAULTS, TIME_DEFAULTS))).format(this);
  });
  define(Date.prototype, 'toLocaleDateString', function toLocaleDateString (locales, options) {
    if (isNaN(this.getTime())) {
      return 'Invalid Date';
    }
    return new (load().DateTimeFormat)(locales, withDefaults(options, DATE_DEFAULTS)).format(this);
  });
  define(Date.prototype, 'toLocaleTimeString', function toLocaleTimeString (locales, options) {
    if (isNaN(this.getTime())) {
      return 'Invalid Date';
    }
    return new (load().DateTimeFormat)(locales, withDefaults(options, TIME_DEFAULTS)).format(this);
  });

  return {};
})
