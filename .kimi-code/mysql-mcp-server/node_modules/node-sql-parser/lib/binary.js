(function (global, factory) {
  if (typeof define === "function" && define.amd) {
    define(["exports", "./expr", "./collate", "./util"], factory);
  } else if (typeof exports !== "undefined") {
    factory(exports, require("./expr"), require("./collate"), require("./util"));
  } else {
    var mod = {
      exports: {}
    };
    factory(mod.exports, global.expr, global.collate, global.util);
    global.binary = mod.exports;
  }
})(typeof globalThis !== "undefined" ? globalThis : typeof self !== "undefined" ? self : this, function (_exports, _expr, _collate, _util) {
  "use strict";

  Object.defineProperty(_exports, "__esModule", {
    value: true
  });
  _exports.binaryToSQL = binaryToSQL;
  function binaryToSQL(expr) {
    let operator = expr.operator || expr.op;
    let rstr = (0, _expr.exprToSQL)(expr.right);
    let isBetween = false;
    if (Array.isArray(rstr)) {
      switch (operator) {
        case '=':
          operator = 'IN';
          break;
        case '!=':
          operator = 'NOT IN';
          break;
        case 'BETWEEN':
        case 'NOT BETWEEN':
          isBetween = true;
          rstr = `${rstr[0]} AND ${rstr[1]}`;
          break;
        default:
          break;
      }
      if (!isBetween) rstr = `(${rstr.join(', ')})`;
    }
    const escape = expr.right.escape || {};
    const leftPart = Array.isArray(expr.left) ? expr.left.map(_expr.exprToSQL).join(', ') : (0, _expr.exprToSQL)(expr.left);
    const collateStr = (0, _collate.collateToSQL)(expr.collate);
    const str = [leftPart, operator, rstr, (0, _util.toUpper)(escape.type), (0, _expr.exprToSQL)(escape.value)].filter(_util.hasVal).join(operator === '.' ? '' : ' ');
    let result = expr.parentheses ? `(${str})` : str;
    if (collateStr) result = `${result} ${collateStr}`;
    return result;
  }
});