var __getOwnPropNames = Object.getOwnPropertyNames;
var __esm = (fn, res, err) => function __init() {
  if (err) throw err[0];
  try {
    return fn && (res = (0, fn[__getOwnPropNames(fn)[0]])(fn = 0)), res;
  } catch (e) {
    throw err = [e], e;
  }
};
var __commonJS = (cb, mod) => function __require() {
  try {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  } catch (e) {
    throw mod = 0, e;
  }
};

// process-shim.js
var queue, process;
var init_process_shim = __esm({
  "process-shim.js"() {
    queue = (fn, ...args) => Promise.resolve().then(() => fn(...args));
    process = {
      env: {},
      argv: [],
      version: "",
      versions: {},
      platform: "browser",
      browser: true,
      title: "browser",
      nextTick: queue,
      cwd: () => "/",
      emitWarning: () => void 0
    };
  }
});

// node_modules/csv-parse/lib/index.js
var require_index = __commonJS({
  "node_modules/csv-parse/lib/index.js"(exports2, module2) {
    init_process_shim();
    var Parser;
    var StringDecoder2;
    var isObjLiteral;
    var stream;
    var util;
    stream = require("stream");
    util = require("util");
    StringDecoder2 = require("string_decoder").StringDecoder;
    module2.exports = function() {
      var callback, called, chunks, data, err, options, parser;
      if (arguments.length === 3) {
        data = arguments[0];
        options = arguments[1];
        callback = arguments[2];
        if (typeof callback !== "function") {
          throw Error("Invalid callback argument: " + JSON.stringify(callback));
        }
        if (!(typeof data === "string" || Buffer.isBuffer(arguments[0]))) {
          return callback(Error("Invalid data argument: " + JSON.stringify(data)));
        }
      } else if (arguments.length === 2) {
        if (typeof arguments[0] === "string" || Buffer.isBuffer(arguments[0])) {
          data = arguments[0];
        } else if (isObjLiteral(arguments[0])) {
          options = arguments[0];
        } else {
          err = "Invalid first argument: " + JSON.stringify(arguments[0]);
        }
        if (typeof arguments[1] === "function") {
          callback = arguments[1];
        } else if (isObjLiteral(arguments[1])) {
          if (options) {
            err = "Invalid arguments: got options twice as first and second arguments";
          } else {
            options = arguments[1];
          }
        } else {
          err = "Invalid first argument: " + JSON.stringify(arguments[1]);
        }
        if (err) {
          if (!callback) {
            throw Error(err);
          } else {
            return callback(Error(err));
          }
        }
      } else if (arguments.length === 1) {
        if (typeof arguments[0] === "function") {
          callback = arguments[0];
        } else {
          options = arguments[0];
        }
      }
      if (options == null) {
        options = {};
      }
      parser = new Parser(options);
      if (data != null) {
        process.nextTick(function() {
          parser.write(data);
          return parser.end();
        });
      }
      if (callback) {
        called = false;
        chunks = options.objname ? {} : [];
        parser.on("readable", function() {
          var chunk, results;
          results = [];
          while (chunk = parser.read()) {
            if (options.objname) {
              results.push(chunks[chunk[0]] = chunk[1]);
            } else {
              results.push(chunks.push(chunk));
            }
          }
          return results;
        });
        parser.on("error", function(err2) {
          called = true;
          return callback(err2);
        });
        parser.on("end", function() {
          if (!called) {
            return callback(null, chunks);
          }
        });
      }
      return parser;
    };
    Parser = function(options) {
      var base, base1, base10, base11, base12, base13, base14, base15, base16, base2, base3, base4, base5, base6, base7, base8, base9, k, v;
      if (options == null) {
        options = {};
      }
      options.objectMode = true;
      this.options = {};
      for (k in options) {
        v = options[k];
        this.options[k] = v;
      }
      stream.Transform.call(this, this.options);
      if ((base = this.options).rowDelimiter == null) {
        base.rowDelimiter = null;
      }
      if (typeof this.options.rowDelimiter === "string") {
        this.options.rowDelimiter = [this.options.rowDelimiter];
      }
      if ((base1 = this.options).delimiter == null) {
        base1.delimiter = ",";
      }
      if (this.options.quote !== void 0 && !this.options.quote) {
        this.options.quote = "";
      }
      if ((base2 = this.options).quote == null) {
        base2.quote = '"';
      }
      if ((base3 = this.options).escape == null) {
        base3.escape = '"';
      }
      if ((base4 = this.options).columns == null) {
        base4.columns = null;
      }
      if ((base5 = this.options).comment == null) {
        base5.comment = "";
      }
      if ((base6 = this.options).objname == null) {
        base6.objname = false;
      }
      if ((base7 = this.options).trim == null) {
        base7.trim = false;
      }
      if ((base8 = this.options).ltrim == null) {
        base8.ltrim = false;
      }
      if ((base9 = this.options).rtrim == null) {
        base9.rtrim = false;
      }
      if ((base10 = this.options).auto_parse == null) {
        base10.auto_parse = false;
      }
      if ((base11 = this.options).auto_parse_date == null) {
        base11.auto_parse_date = false;
      }
      if ((base12 = this.options).relax == null) {
        base12.relax = false;
      }
      if ((base13 = this.options).relax_column_count == null) {
        base13.relax_column_count = false;
      }
      if ((base14 = this.options).skip_empty_lines == null) {
        base14.skip_empty_lines = false;
      }
      if ((base15 = this.options).max_limit_on_data_read == null) {
        base15.max_limit_on_data_read = 128e3;
      }
      if ((base16 = this.options).skip_lines_with_empty_values == null) {
        base16.skip_lines_with_empty_values = false;
      }
      this.lines = 0;
      this.count = 0;
      this.skipped_line_count = 0;
      this.empty_line_count = 0;
      this.is_int = /^(\-|\+)?([1-9]+[0-9]*)$/;
      this.is_float = function(value) {
        return value - parseFloat(value) + 1 >= 0;
      };
      this._ = {
        decoder: new StringDecoder2(),
        quoting: false,
        commenting: false,
        field: null,
        nextChar: null,
        closingQuote: 0,
        line: [],
        chunks: [],
        rawBuf: "",
        buf: "",
        rowDelimiterLength: this.options.rowDelimiter ? Math.max.apply(Math, this.options.rowDelimiter.map(function(v2) {
          return v2.length;
        })) : void 0
      };
      return this;
    };
    util.inherits(Parser, stream.Transform);
    module2.exports.Parser = Parser;
    Parser.prototype._transform = function(chunk, encoding, callback) {
      var err;
      if (chunk instanceof Buffer) {
        chunk = this._.decoder.write(chunk);
      }
      err = this.__write(chunk, false);
      if (err) {
        return this.emit("error", err);
      }
      return callback();
    };
    Parser.prototype._flush = function(callback) {
      var err;
      err = this.__write(this._.decoder.end(), true);
      if (err) {
        return this.emit("error", err);
      }
      if (this._.quoting) {
        this.emit("error", new Error("Quoted field not terminated at line " + (this.lines + 1)));
        return;
      }
      if (this._.line.length > 0) {
        err = this.__push(this._.line);
        if (err) {
          return callback(err);
        }
      }
      return callback();
    };
    Parser.prototype.__push = function(line) {
      var call_column_udf, columns, err, field, i, j, len, lineAsColumns, rawBuf2, ref, row;
      if (this.options.skip_lines_with_empty_values && line.join("").trim() === "") {
        return;
      }
      row = null;
      if (this.options.columns === true) {
        this.options.columns = line;
        rawBuf2 = "";
        return;
      } else if (typeof this.options.columns === "function") {
        call_column_udf = function(fn, line2) {
          var columns2, err2;
          try {
            columns2 = fn.call(null, line2);
            return [null, columns2];
          } catch (error) {
            err2 = error;
            return [err2];
          }
        };
        ref = call_column_udf(this.options.columns, line), err = ref[0], columns = ref[1];
        if (err) {
          return err;
        }
        this.options.columns = columns;
        rawBuf2 = "";
        return;
      }
      if (!this._.line_length && line.length > 0) {
        this._.line_length = this.options.columns ? this.options.columns.length : line.length;
      }
      if (line.length === 1 && line[0] === "") {
        this.empty_line_count++;
      } else if (line.length !== this._.line_length) {
        if (this.options.relax_column_count) {
          this.count++;
          this.skipped_line_count++;
        } else if (this.options.columns != null) {
          return Error("Number of columns on line " + this.lines + " does not match header");
        } else {
          return Error("Number of columns is inconsistent on line " + this.lines);
        }
      } else {
        this.count++;
      }
      if (this.options.columns != null) {
        lineAsColumns = {};
        for (i = j = 0, len = line.length; j < len; i = ++j) {
          field = line[i];
          if (this.options.columns[i] === false) {
            continue;
          }
          lineAsColumns[this.options.columns[i]] = field;
        }
        if (this.options.objname) {
          row = [lineAsColumns[this.options.objname], lineAsColumns];
        } else {
          row = lineAsColumns;
        }
      } else {
        row = line;
      }
      if (this.count < this.options.from) {
        return;
      }
      if (this.count > this.options.to) {
        return;
      }
      if (this.options.raw) {
        this.push({
          raw: this._.rawBuf,
          row
        });
        this._.rawBuf = "";
      } else {
        this.push(row);
      }
      return null;
    };
    Parser.prototype.__write = function(chars, end) {
      var areNextCharsDelimiter, areNextCharsRowDelimiters, auto_parse, char, err, escapeIsQuote, i, isDelimiter, isEscape, isNextCharAComment, isQuote, isRowDelimiter, isRowDelimiterLength, is_float, is_int, l, ltrim, nextCharPos, ref, ref1, ref2, ref3, ref4, ref5, remainingBuffer, rowDelimiter, rtrim, wasCommenting;
      is_int = /* @__PURE__ */ (function(_this) {
        return function(value) {
          if (typeof _this.is_int === "function") {
            return _this.is_int(value);
          } else {
            return _this.is_int.test(value);
          }
        };
      })(this);
      is_float = /* @__PURE__ */ (function(_this) {
        return function(value) {
          if (typeof _this.is_float === "function") {
            return _this.is_float(value);
          } else {
            return _this.is_float.test(value);
          }
        };
      })(this);
      auto_parse = /* @__PURE__ */ (function(_this) {
        return function(value) {
          var m;
          if (!_this.options.auto_parse) {
            return value;
          }
          if (is_int(value)) {
            value = parseInt(value);
          } else if (is_float(value)) {
            value = parseFloat(value);
          } else if (_this.options.auto_parse_date) {
            m = Date.parse(value);
            if (!isNaN(m)) {
              value = new Date(m);
            }
          }
          return value;
        };
      })(this);
      ltrim = this.options.trim || this.options.ltrim;
      rtrim = this.options.trim || this.options.rtrim;
      chars = this._.buf + chars;
      l = chars.length;
      i = 0;
      if (this.lines === 0 && 65279 === chars.charCodeAt(0)) {
        i++;
      }
      while (i < l) {
        if (!end) {
          remainingBuffer = chars.substr(i, l - i);
          if (!this.options.rowDelimiter && i + 3 > l || !this._.commenting && l - i < this.options.comment.length && this.options.comment.substr(0, l - i) === remainingBuffer || this.options.rowDelimiter && l - i < this._.rowDelimiterLength && this.options.rowDelimiter.some(function(rd) {
            return rd.substr(0, l - i) === remainingBuffer;
          }) || this.options.rowDelimiter && this._.quoting && l - i < this.options.quote.length + this._.rowDelimiterLength && this.options.rowDelimiter.some(/* @__PURE__ */ (function(_this) {
            return function(rd) {
              return (_this.options.quote + rd).substr(0, l - i) === remainingBuffer;
            };
          })(this)) || l - i <= this.options.delimiter.length && this.options.delimiter.substr(0, l - i) === remainingBuffer || l - i <= this.options.escape.length && this.options.escape.substr(0, l - i) === remainingBuffer) {
            break;
          }
        }
        char = this._.nextChar ? this._.nextChar : chars.charAt(i);
        this._.nextChar = l > i + 1 ? chars.charAt(i + 1) : "";
        if (this.options.raw) {
          this._.rawBuf += char;
        }
        if (this.options.rowDelimiter == null) {
          nextCharPos = i;
          rowDelimiter = null;
          if (!this._.quoting && (char === "\n" || char === "\r")) {
            rowDelimiter = char;
            nextCharPos += 1;
          } else if (this._.quoting && char === this.options.quote && ((ref = this._.nextChar) === "\n" || ref === "\r")) {
            rowDelimiter = this._.nextChar;
            nextCharPos += 2;
            if (this.raw) {
              rawBuf += this._.nextChar;
            }
          }
          if (rowDelimiter) {
            if (rowDelimiter === "\r" && chars.charAt(nextCharPos) === "\n") {
              rowDelimiter += "\n";
            }
            this.options.rowDelimiter = [rowDelimiter];
            this._.rowDelimiterLength = rowDelimiter.length;
          }
        }
        if (!this._.commenting && char === this.options.escape) {
          escapeIsQuote = this.options.escape === this.options.quote;
          isEscape = this._.nextChar === this.options.escape;
          isQuote = this._.nextChar === this.options.quote;
          if (!(escapeIsQuote && this._.field == null && !this._.quoting) && (isEscape || isQuote)) {
            i++;
            char = this._.nextChar;
            this._.nextChar = chars.charAt(i + 1);
            if (this._.field == null) {
              this._.field = "";
            }
            this._.field += char;
            if (this.options.raw) {
              this._.rawBuf += char;
            }
            i++;
            continue;
          }
        }
        if (!this._.commenting && char === this.options.quote) {
          if (this._.quoting) {
            areNextCharsRowDelimiters = this.options.rowDelimiter && this.options.rowDelimiter.some(function(rd) {
              return chars.substr(i + 1, rd.length) === rd;
            });
            areNextCharsDelimiter = chars.substr(i + 1, this.options.delimiter.length) === this.options.delimiter;
            isNextCharAComment = this._.nextChar === this.options.comment;
            if (this._.nextChar && !areNextCharsRowDelimiters && !areNextCharsDelimiter && !isNextCharAComment) {
              if (this.options.relax) {
                this._.quoting = false;
                if (this._.field) {
                  this._.field = "" + this.options.quote + this._.field;
                }
              } else {
                return Error("Invalid closing quote at line " + (this.lines + 1) + "; found " + JSON.stringify(this._.nextChar) + " instead of delimiter " + JSON.stringify(this.options.delimiter));
              }
            } else {
              this._.quoting = false;
              this._.closingQuote = this.options.quote.length;
              i++;
              if (end && i === l) {
                this._.line.push(auto_parse(this._.field || ""));
                this._.field = null;
              }
              continue;
            }
          } else if (!this._.field) {
            this._.quoting = true;
            i++;
            continue;
          } else if (this._.field != null && !this.options.relax) {
            return Error("Invalid opening quote at line " + (this.lines + 1));
          }
        }
        isRowDelimiter = this.options.rowDelimiter && this.options.rowDelimiter.some(function(rd) {
          return chars.substr(i, rd.length) === rd;
        });
        if (isRowDelimiter || end && i === l - 1) {
          this.lines++;
        }
        wasCommenting = false;
        if (!this._.commenting && !this._.quoting && this.options.comment && chars.substr(i, this.options.comment.length) === this.options.comment) {
          this._.commenting = true;
        } else if (this._.commenting && isRowDelimiter) {
          wasCommenting = true;
          this._.commenting = false;
        }
        isDelimiter = chars.substr(i, this.options.delimiter.length) === this.options.delimiter;
        if (!this._.commenting && !this._.quoting && (isDelimiter || isRowDelimiter)) {
          if (isRowDelimiter) {
            isRowDelimiterLength = this.options.rowDelimiter.filter(function(rd) {
              return chars.substr(i, rd.length) === rd;
            })[0].length;
          }
          if (isRowDelimiter && this._.line.length === 0 && this._.field == null) {
            if (wasCommenting || this.options.skip_empty_lines) {
              i += isRowDelimiterLength;
              this._.nextChar = chars.charAt(i);
              continue;
            }
          }
          if (rtrim) {
            if (!this._.closingQuote) {
              this._.field = (ref1 = this._.field) != null ? ref1.trimRight() : void 0;
            }
          }
          this._.line.push(auto_parse(this._.field || ""));
          this._.closingQuote = 0;
          this._.field = null;
          if (isDelimiter) {
            i += this.options.delimiter.length;
            this._.nextChar = chars.charAt(i);
            if (end && !this._.nextChar) {
              isRowDelimiter = true;
              this._.line.push("");
            }
          }
          if (isRowDelimiter) {
            err = this.__push(this._.line);
            if (err) {
              return err;
            }
            this._.line = [];
            i += isRowDelimiterLength;
            this._.nextChar = chars.charAt(i);
            continue;
          }
        } else if (!this._.commenting && !this._.quoting && (char === " " || char === "	")) {
          if (this._.field == null) {
            this._.field = "";
          }
          if (!(ltrim && !this._.field)) {
            this._.field += char;
          }
          i++;
        } else if (!this._.commenting) {
          if (this._.field == null) {
            this._.field = "";
          }
          this._.field += char;
          i++;
        } else {
          i++;
        }
        if (!this._.commenting && ((ref2 = this._.field) != null ? ref2.length : void 0) > this.options.max_limit_on_data_read) {
          return Error("Field exceeds max_limit_on_data_read setting (" + this.options.max_limit_on_data_read + ") " + JSON.stringify(this.options.delimiter));
        }
        if (!this._.commenting && ((ref3 = this._.line) != null ? ref3.length : void 0) > this.options.max_limit_on_data_read) {
          return Error("Row delimiter not found in the file " + JSON.stringify(this.options.rowDelimiter));
        }
      }
      if (end) {
        if (this._.field != null) {
          if (rtrim) {
            if (!this._.closingQuote) {
              this._.field = (ref4 = this._.field) != null ? ref4.trimRight() : void 0;
            }
          }
          this._.line.push(auto_parse(this._.field || ""));
          this._.field = null;
        }
        if (((ref5 = this._.field) != null ? ref5.length : void 0) > this.options.max_limit_on_data_read) {
          return Error("Delimiter not found in the file " + JSON.stringify(this.options.delimiter));
        }
        if (l === 0) {
          this.lines++;
        }
        if (this._.line.length > this.options.max_limit_on_data_read) {
          return Error("Row delimiter not found in the file " + JSON.stringify(this.options.rowDelimiter));
        }
      }
      this._.buf = chars.substr(i);
      return null;
    };
    isObjLiteral = function(_obj) {
      var _test;
      _test = _obj;
      if (typeof _obj !== "object" || _obj === null || Array.isArray(_obj)) {
        return false;
      } else {
        return (function() {
          while (true) {
            if (Object.getPrototypeOf(_test = Object.getPrototypeOf(_test)) === null) {
              break;
            }
          }
          return Object.getPrototypeOf(_obj === _test);
        })();
      }
    };
  }
});

// node_modules/csv-parse/lib/sync.js
init_process_shim();
var StringDecoder;
var parse;
StringDecoder = require("string_decoder").StringDecoder;
parse = require_index();
module.exports = function(data, options) {
  var decoder, err, parser, records;
  if (options == null) {
    options = {};
  }
  records = options.objname ? {} : [];
  if (data instanceof Buffer) {
    decoder = new StringDecoder();
    data = decoder.write(data);
  }
  parser = new parse.Parser(options);
  parser.push = function(record) {
    if (options.objname) {
      return records[record[0]] = record[1];
    } else {
      return records.push(record);
    }
  };
  err = parser.__write(data, false);
  if (err) {
    throw err;
  }
  if (data instanceof Buffer) {
    err = parser.__write(data.end(), true);
    if (err) {
      throw err;
    }
  }
  parser._flush((function() {
  }));
  return records;
};
