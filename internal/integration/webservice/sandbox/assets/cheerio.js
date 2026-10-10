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
var init_process_shim = __esm({
  "process-shim.js"() {
  }
});

// node_modules/entities/maps/decode.json
var require_decode = __commonJS({
  "node_modules/entities/maps/decode.json"(exports2, module2) {
    module2.exports = { "0": 65533, "128": 8364, "130": 8218, "131": 402, "132": 8222, "133": 8230, "134": 8224, "135": 8225, "136": 710, "137": 8240, "138": 352, "139": 8249, "140": 338, "142": 381, "145": 8216, "146": 8217, "147": 8220, "148": 8221, "149": 8226, "150": 8211, "151": 8212, "152": 732, "153": 8482, "154": 353, "155": 8250, "156": 339, "158": 382, "159": 376 };
  }
});

// node_modules/entities/lib/decode_codepoint.js
var require_decode_codepoint = __commonJS({
  "node_modules/entities/lib/decode_codepoint.js"(exports2, module2) {
    init_process_shim();
    var decodeMap = require_decode();
    module2.exports = decodeCodePoint;
    function decodeCodePoint(codePoint) {
      if (codePoint >= 55296 && codePoint <= 57343 || codePoint > 1114111) {
        return "�";
      }
      if (codePoint in decodeMap) {
        codePoint = decodeMap[codePoint];
      }
      var output = "";
      if (codePoint > 65535) {
        codePoint -= 65536;
        output += String.fromCharCode(codePoint >>> 10 & 1023 | 55296);
        codePoint = 56320 | codePoint & 1023;
      }
      output += String.fromCharCode(codePoint);
      return output;
    }
  }
});

// node_modules/entities/maps/entities.json
var require_entities = __commonJS({
  "node_modules/entities/maps/entities.json"(exports2, module2) {
    module2.exports = { Aacute: "Á", aacute: "á", Abreve: "Ă", abreve: "ă", ac: "∾", acd: "∿", acE: "∾̳", Acirc: "Â", acirc: "â", acute: "´", Acy: "А", acy: "а", AElig: "Æ", aelig: "æ", af: "⁡", Afr: "𝔄", afr: "𝔞", Agrave: "À", agrave: "à", alefsym: "ℵ", aleph: "ℵ", Alpha: "Α", alpha: "α", Amacr: "Ā", amacr: "ā", amalg: "⨿", amp: "&", AMP: "&", andand: "⩕", And: "⩓", and: "∧", andd: "⩜", andslope: "⩘", andv: "⩚", ang: "∠", ange: "⦤", angle: "∠", angmsdaa: "⦨", angmsdab: "⦩", angmsdac: "⦪", angmsdad: "⦫", angmsdae: "⦬", angmsdaf: "⦭", angmsdag: "⦮", angmsdah: "⦯", angmsd: "∡", angrt: "∟", angrtvb: "⊾", angrtvbd: "⦝", angsph: "∢", angst: "Å", angzarr: "⍼", Aogon: "Ą", aogon: "ą", Aopf: "𝔸", aopf: "𝕒", apacir: "⩯", ap: "≈", apE: "⩰", ape: "≊", apid: "≋", apos: "'", ApplyFunction: "⁡", approx: "≈", approxeq: "≊", Aring: "Å", aring: "å", Ascr: "𝒜", ascr: "𝒶", Assign: "≔", ast: "*", asymp: "≈", asympeq: "≍", Atilde: "Ã", atilde: "ã", Auml: "Ä", auml: "ä", awconint: "∳", awint: "⨑", backcong: "≌", backepsilon: "϶", backprime: "‵", backsim: "∽", backsimeq: "⋍", Backslash: "∖", Barv: "⫧", barvee: "⊽", barwed: "⌅", Barwed: "⌆", barwedge: "⌅", bbrk: "⎵", bbrktbrk: "⎶", bcong: "≌", Bcy: "Б", bcy: "б", bdquo: "„", becaus: "∵", because: "∵", Because: "∵", bemptyv: "⦰", bepsi: "϶", bernou: "ℬ", Bernoullis: "ℬ", Beta: "Β", beta: "β", beth: "ℶ", between: "≬", Bfr: "𝔅", bfr: "𝔟", bigcap: "⋂", bigcirc: "◯", bigcup: "⋃", bigodot: "⨀", bigoplus: "⨁", bigotimes: "⨂", bigsqcup: "⨆", bigstar: "★", bigtriangledown: "▽", bigtriangleup: "△", biguplus: "⨄", bigvee: "⋁", bigwedge: "⋀", bkarow: "⤍", blacklozenge: "⧫", blacksquare: "▪", blacktriangle: "▴", blacktriangledown: "▾", blacktriangleleft: "◂", blacktriangleright: "▸", blank: "␣", blk12: "▒", blk14: "░", blk34: "▓", block: "█", bne: "=⃥", bnequiv: "≡⃥", bNot: "⫭", bnot: "⌐", Bopf: "𝔹", bopf: "𝕓", bot: "⊥", bottom: "⊥", bowtie: "⋈", boxbox: "⧉", boxdl: "┐", boxdL: "╕", boxDl: "╖", boxDL: "╗", boxdr: "┌", boxdR: "╒", boxDr: "╓", boxDR: "╔", boxh: "─", boxH: "═", boxhd: "┬", boxHd: "╤", boxhD: "╥", boxHD: "╦", boxhu: "┴", boxHu: "╧", boxhU: "╨", boxHU: "╩", boxminus: "⊟", boxplus: "⊞", boxtimes: "⊠", boxul: "┘", boxuL: "╛", boxUl: "╜", boxUL: "╝", boxur: "└", boxuR: "╘", boxUr: "╙", boxUR: "╚", boxv: "│", boxV: "║", boxvh: "┼", boxvH: "╪", boxVh: "╫", boxVH: "╬", boxvl: "┤", boxvL: "╡", boxVl: "╢", boxVL: "╣", boxvr: "├", boxvR: "╞", boxVr: "╟", boxVR: "╠", bprime: "‵", breve: "˘", Breve: "˘", brvbar: "¦", bscr: "𝒷", Bscr: "ℬ", bsemi: "⁏", bsim: "∽", bsime: "⋍", bsolb: "⧅", bsol: "\\", bsolhsub: "⟈", bull: "•", bullet: "•", bump: "≎", bumpE: "⪮", bumpe: "≏", Bumpeq: "≎", bumpeq: "≏", Cacute: "Ć", cacute: "ć", capand: "⩄", capbrcup: "⩉", capcap: "⩋", cap: "∩", Cap: "⋒", capcup: "⩇", capdot: "⩀", CapitalDifferentialD: "ⅅ", caps: "∩︀", caret: "⁁", caron: "ˇ", Cayleys: "ℭ", ccaps: "⩍", Ccaron: "Č", ccaron: "č", Ccedil: "Ç", ccedil: "ç", Ccirc: "Ĉ", ccirc: "ĉ", Cconint: "∰", ccups: "⩌", ccupssm: "⩐", Cdot: "Ċ", cdot: "ċ", cedil: "¸", Cedilla: "¸", cemptyv: "⦲", cent: "¢", centerdot: "·", CenterDot: "·", cfr: "𝔠", Cfr: "ℭ", CHcy: "Ч", chcy: "ч", check: "✓", checkmark: "✓", Chi: "Χ", chi: "χ", circ: "ˆ", circeq: "≗", circlearrowleft: "↺", circlearrowright: "↻", circledast: "⊛", circledcirc: "⊚", circleddash: "⊝", CircleDot: "⊙", circledR: "®", circledS: "Ⓢ", CircleMinus: "⊖", CirclePlus: "⊕", CircleTimes: "⊗", cir: "○", cirE: "⧃", cire: "≗", cirfnint: "⨐", cirmid: "⫯", cirscir: "⧂", ClockwiseContourIntegral: "∲", CloseCurlyDoubleQuote: "”", CloseCurlyQuote: "’", clubs: "♣", clubsuit: "♣", colon: ":", Colon: "∷", Colone: "⩴", colone: "≔", coloneq: "≔", comma: ",", commat: "@", comp: "∁", compfn: "∘", complement: "∁", complexes: "ℂ", cong: "≅", congdot: "⩭", Congruent: "≡", conint: "∮", Conint: "∯", ContourIntegral: "∮", copf: "𝕔", Copf: "ℂ", coprod: "∐", Coproduct: "∐", copy: "©", COPY: "©", copysr: "℗", CounterClockwiseContourIntegral: "∳", crarr: "↵", cross: "✗", Cross: "⨯", Cscr: "𝒞", cscr: "𝒸", csub: "⫏", csube: "⫑", csup: "⫐", csupe: "⫒", ctdot: "⋯", cudarrl: "⤸", cudarrr: "⤵", cuepr: "⋞", cuesc: "⋟", cularr: "↶", cularrp: "⤽", cupbrcap: "⩈", cupcap: "⩆", CupCap: "≍", cup: "∪", Cup: "⋓", cupcup: "⩊", cupdot: "⊍", cupor: "⩅", cups: "∪︀", curarr: "↷", curarrm: "⤼", curlyeqprec: "⋞", curlyeqsucc: "⋟", curlyvee: "⋎", curlywedge: "⋏", curren: "¤", curvearrowleft: "↶", curvearrowright: "↷", cuvee: "⋎", cuwed: "⋏", cwconint: "∲", cwint: "∱", cylcty: "⌭", dagger: "†", Dagger: "‡", daleth: "ℸ", darr: "↓", Darr: "↡", dArr: "⇓", dash: "‐", Dashv: "⫤", dashv: "⊣", dbkarow: "⤏", dblac: "˝", Dcaron: "Ď", dcaron: "ď", Dcy: "Д", dcy: "д", ddagger: "‡", ddarr: "⇊", DD: "ⅅ", dd: "ⅆ", DDotrahd: "⤑", ddotseq: "⩷", deg: "°", Del: "∇", Delta: "Δ", delta: "δ", demptyv: "⦱", dfisht: "⥿", Dfr: "𝔇", dfr: "𝔡", dHar: "⥥", dharl: "⇃", dharr: "⇂", DiacriticalAcute: "´", DiacriticalDot: "˙", DiacriticalDoubleAcute: "˝", DiacriticalGrave: "`", DiacriticalTilde: "˜", diam: "⋄", diamond: "⋄", Diamond: "⋄", diamondsuit: "♦", diams: "♦", die: "¨", DifferentialD: "ⅆ", digamma: "ϝ", disin: "⋲", div: "÷", divide: "÷", divideontimes: "⋇", divonx: "⋇", DJcy: "Ђ", djcy: "ђ", dlcorn: "⌞", dlcrop: "⌍", dollar: "$", Dopf: "𝔻", dopf: "𝕕", Dot: "¨", dot: "˙", DotDot: "⃜", doteq: "≐", doteqdot: "≑", DotEqual: "≐", dotminus: "∸", dotplus: "∔", dotsquare: "⊡", doublebarwedge: "⌆", DoubleContourIntegral: "∯", DoubleDot: "¨", DoubleDownArrow: "⇓", DoubleLeftArrow: "⇐", DoubleLeftRightArrow: "⇔", DoubleLeftTee: "⫤", DoubleLongLeftArrow: "⟸", DoubleLongLeftRightArrow: "⟺", DoubleLongRightArrow: "⟹", DoubleRightArrow: "⇒", DoubleRightTee: "⊨", DoubleUpArrow: "⇑", DoubleUpDownArrow: "⇕", DoubleVerticalBar: "∥", DownArrowBar: "⤓", downarrow: "↓", DownArrow: "↓", Downarrow: "⇓", DownArrowUpArrow: "⇵", DownBreve: "̑", downdownarrows: "⇊", downharpoonleft: "⇃", downharpoonright: "⇂", DownLeftRightVector: "⥐", DownLeftTeeVector: "⥞", DownLeftVectorBar: "⥖", DownLeftVector: "↽", DownRightTeeVector: "⥟", DownRightVectorBar: "⥗", DownRightVector: "⇁", DownTeeArrow: "↧", DownTee: "⊤", drbkarow: "⤐", drcorn: "⌟", drcrop: "⌌", Dscr: "𝒟", dscr: "𝒹", DScy: "Ѕ", dscy: "ѕ", dsol: "⧶", Dstrok: "Đ", dstrok: "đ", dtdot: "⋱", dtri: "▿", dtrif: "▾", duarr: "⇵", duhar: "⥯", dwangle: "⦦", DZcy: "Џ", dzcy: "џ", dzigrarr: "⟿", Eacute: "É", eacute: "é", easter: "⩮", Ecaron: "Ě", ecaron: "ě", Ecirc: "Ê", ecirc: "ê", ecir: "≖", ecolon: "≕", Ecy: "Э", ecy: "э", eDDot: "⩷", Edot: "Ė", edot: "ė", eDot: "≑", ee: "ⅇ", efDot: "≒", Efr: "𝔈", efr: "𝔢", eg: "⪚", Egrave: "È", egrave: "è", egs: "⪖", egsdot: "⪘", el: "⪙", Element: "∈", elinters: "⏧", ell: "ℓ", els: "⪕", elsdot: "⪗", Emacr: "Ē", emacr: "ē", empty: "∅", emptyset: "∅", EmptySmallSquare: "◻", emptyv: "∅", EmptyVerySmallSquare: "▫", emsp13: " ", emsp14: " ", emsp: " ", ENG: "Ŋ", eng: "ŋ", ensp: " ", Eogon: "Ę", eogon: "ę", Eopf: "𝔼", eopf: "𝕖", epar: "⋕", eparsl: "⧣", eplus: "⩱", epsi: "ε", Epsilon: "Ε", epsilon: "ε", epsiv: "ϵ", eqcirc: "≖", eqcolon: "≕", eqsim: "≂", eqslantgtr: "⪖", eqslantless: "⪕", Equal: "⩵", equals: "=", EqualTilde: "≂", equest: "≟", Equilibrium: "⇌", equiv: "≡", equivDD: "⩸", eqvparsl: "⧥", erarr: "⥱", erDot: "≓", escr: "ℯ", Escr: "ℰ", esdot: "≐", Esim: "⩳", esim: "≂", Eta: "Η", eta: "η", ETH: "Ð", eth: "ð", Euml: "Ë", euml: "ë", euro: "€", excl: "!", exist: "∃", Exists: "∃", expectation: "ℰ", exponentiale: "ⅇ", ExponentialE: "ⅇ", fallingdotseq: "≒", Fcy: "Ф", fcy: "ф", female: "♀", ffilig: "ﬃ", fflig: "ﬀ", ffllig: "ﬄ", Ffr: "𝔉", ffr: "𝔣", filig: "ﬁ", FilledSmallSquare: "◼", FilledVerySmallSquare: "▪", fjlig: "fj", flat: "♭", fllig: "ﬂ", fltns: "▱", fnof: "ƒ", Fopf: "𝔽", fopf: "𝕗", forall: "∀", ForAll: "∀", fork: "⋔", forkv: "⫙", Fouriertrf: "ℱ", fpartint: "⨍", frac12: "½", frac13: "⅓", frac14: "¼", frac15: "⅕", frac16: "⅙", frac18: "⅛", frac23: "⅔", frac25: "⅖", frac34: "¾", frac35: "⅗", frac38: "⅜", frac45: "⅘", frac56: "⅚", frac58: "⅝", frac78: "⅞", frasl: "⁄", frown: "⌢", fscr: "𝒻", Fscr: "ℱ", gacute: "ǵ", Gamma: "Γ", gamma: "γ", Gammad: "Ϝ", gammad: "ϝ", gap: "⪆", Gbreve: "Ğ", gbreve: "ğ", Gcedil: "Ģ", Gcirc: "Ĝ", gcirc: "ĝ", Gcy: "Г", gcy: "г", Gdot: "Ġ", gdot: "ġ", ge: "≥", gE: "≧", gEl: "⪌", gel: "⋛", geq: "≥", geqq: "≧", geqslant: "⩾", gescc: "⪩", ges: "⩾", gesdot: "⪀", gesdoto: "⪂", gesdotol: "⪄", gesl: "⋛︀", gesles: "⪔", Gfr: "𝔊", gfr: "𝔤", gg: "≫", Gg: "⋙", ggg: "⋙", gimel: "ℷ", GJcy: "Ѓ", gjcy: "ѓ", gla: "⪥", gl: "≷", glE: "⪒", glj: "⪤", gnap: "⪊", gnapprox: "⪊", gne: "⪈", gnE: "≩", gneq: "⪈", gneqq: "≩", gnsim: "⋧", Gopf: "𝔾", gopf: "𝕘", grave: "`", GreaterEqual: "≥", GreaterEqualLess: "⋛", GreaterFullEqual: "≧", GreaterGreater: "⪢", GreaterLess: "≷", GreaterSlantEqual: "⩾", GreaterTilde: "≳", Gscr: "𝒢", gscr: "ℊ", gsim: "≳", gsime: "⪎", gsiml: "⪐", gtcc: "⪧", gtcir: "⩺", gt: ">", GT: ">", Gt: "≫", gtdot: "⋗", gtlPar: "⦕", gtquest: "⩼", gtrapprox: "⪆", gtrarr: "⥸", gtrdot: "⋗", gtreqless: "⋛", gtreqqless: "⪌", gtrless: "≷", gtrsim: "≳", gvertneqq: "≩︀", gvnE: "≩︀", Hacek: "ˇ", hairsp: " ", half: "½", hamilt: "ℋ", HARDcy: "Ъ", hardcy: "ъ", harrcir: "⥈", harr: "↔", hArr: "⇔", harrw: "↭", Hat: "^", hbar: "ℏ", Hcirc: "Ĥ", hcirc: "ĥ", hearts: "♥", heartsuit: "♥", hellip: "…", hercon: "⊹", hfr: "𝔥", Hfr: "ℌ", HilbertSpace: "ℋ", hksearow: "⤥", hkswarow: "⤦", hoarr: "⇿", homtht: "∻", hookleftarrow: "↩", hookrightarrow: "↪", hopf: "𝕙", Hopf: "ℍ", horbar: "―", HorizontalLine: "─", hscr: "𝒽", Hscr: "ℋ", hslash: "ℏ", Hstrok: "Ħ", hstrok: "ħ", HumpDownHump: "≎", HumpEqual: "≏", hybull: "⁃", hyphen: "‐", Iacute: "Í", iacute: "í", ic: "⁣", Icirc: "Î", icirc: "î", Icy: "И", icy: "и", Idot: "İ", IEcy: "Е", iecy: "е", iexcl: "¡", iff: "⇔", ifr: "𝔦", Ifr: "ℑ", Igrave: "Ì", igrave: "ì", ii: "ⅈ", iiiint: "⨌", iiint: "∭", iinfin: "⧜", iiota: "℩", IJlig: "Ĳ", ijlig: "ĳ", Imacr: "Ī", imacr: "ī", image: "ℑ", ImaginaryI: "ⅈ", imagline: "ℐ", imagpart: "ℑ", imath: "ı", Im: "ℑ", imof: "⊷", imped: "Ƶ", Implies: "⇒", incare: "℅", in: "∈", infin: "∞", infintie: "⧝", inodot: "ı", intcal: "⊺", int: "∫", Int: "∬", integers: "ℤ", Integral: "∫", intercal: "⊺", Intersection: "⋂", intlarhk: "⨗", intprod: "⨼", InvisibleComma: "⁣", InvisibleTimes: "⁢", IOcy: "Ё", iocy: "ё", Iogon: "Į", iogon: "į", Iopf: "𝕀", iopf: "𝕚", Iota: "Ι", iota: "ι", iprod: "⨼", iquest: "¿", iscr: "𝒾", Iscr: "ℐ", isin: "∈", isindot: "⋵", isinE: "⋹", isins: "⋴", isinsv: "⋳", isinv: "∈", it: "⁢", Itilde: "Ĩ", itilde: "ĩ", Iukcy: "І", iukcy: "і", Iuml: "Ï", iuml: "ï", Jcirc: "Ĵ", jcirc: "ĵ", Jcy: "Й", jcy: "й", Jfr: "𝔍", jfr: "𝔧", jmath: "ȷ", Jopf: "𝕁", jopf: "𝕛", Jscr: "𝒥", jscr: "𝒿", Jsercy: "Ј", jsercy: "ј", Jukcy: "Є", jukcy: "є", Kappa: "Κ", kappa: "κ", kappav: "ϰ", Kcedil: "Ķ", kcedil: "ķ", Kcy: "К", kcy: "к", Kfr: "𝔎", kfr: "𝔨", kgreen: "ĸ", KHcy: "Х", khcy: "х", KJcy: "Ќ", kjcy: "ќ", Kopf: "𝕂", kopf: "𝕜", Kscr: "𝒦", kscr: "𝓀", lAarr: "⇚", Lacute: "Ĺ", lacute: "ĺ", laemptyv: "⦴", lagran: "ℒ", Lambda: "Λ", lambda: "λ", lang: "⟨", Lang: "⟪", langd: "⦑", langle: "⟨", lap: "⪅", Laplacetrf: "ℒ", laquo: "«", larrb: "⇤", larrbfs: "⤟", larr: "←", Larr: "↞", lArr: "⇐", larrfs: "⤝", larrhk: "↩", larrlp: "↫", larrpl: "⤹", larrsim: "⥳", larrtl: "↢", latail: "⤙", lAtail: "⤛", lat: "⪫", late: "⪭", lates: "⪭︀", lbarr: "⤌", lBarr: "⤎", lbbrk: "❲", lbrace: "{", lbrack: "[", lbrke: "⦋", lbrksld: "⦏", lbrkslu: "⦍", Lcaron: "Ľ", lcaron: "ľ", Lcedil: "Ļ", lcedil: "ļ", lceil: "⌈", lcub: "{", Lcy: "Л", lcy: "л", ldca: "⤶", ldquo: "“", ldquor: "„", ldrdhar: "⥧", ldrushar: "⥋", ldsh: "↲", le: "≤", lE: "≦", LeftAngleBracket: "⟨", LeftArrowBar: "⇤", leftarrow: "←", LeftArrow: "←", Leftarrow: "⇐", LeftArrowRightArrow: "⇆", leftarrowtail: "↢", LeftCeiling: "⌈", LeftDoubleBracket: "⟦", LeftDownTeeVector: "⥡", LeftDownVectorBar: "⥙", LeftDownVector: "⇃", LeftFloor: "⌊", leftharpoondown: "↽", leftharpoonup: "↼", leftleftarrows: "⇇", leftrightarrow: "↔", LeftRightArrow: "↔", Leftrightarrow: "⇔", leftrightarrows: "⇆", leftrightharpoons: "⇋", leftrightsquigarrow: "↭", LeftRightVector: "⥎", LeftTeeArrow: "↤", LeftTee: "⊣", LeftTeeVector: "⥚", leftthreetimes: "⋋", LeftTriangleBar: "⧏", LeftTriangle: "⊲", LeftTriangleEqual: "⊴", LeftUpDownVector: "⥑", LeftUpTeeVector: "⥠", LeftUpVectorBar: "⥘", LeftUpVector: "↿", LeftVectorBar: "⥒", LeftVector: "↼", lEg: "⪋", leg: "⋚", leq: "≤", leqq: "≦", leqslant: "⩽", lescc: "⪨", les: "⩽", lesdot: "⩿", lesdoto: "⪁", lesdotor: "⪃", lesg: "⋚︀", lesges: "⪓", lessapprox: "⪅", lessdot: "⋖", lesseqgtr: "⋚", lesseqqgtr: "⪋", LessEqualGreater: "⋚", LessFullEqual: "≦", LessGreater: "≶", lessgtr: "≶", LessLess: "⪡", lesssim: "≲", LessSlantEqual: "⩽", LessTilde: "≲", lfisht: "⥼", lfloor: "⌊", Lfr: "𝔏", lfr: "𝔩", lg: "≶", lgE: "⪑", lHar: "⥢", lhard: "↽", lharu: "↼", lharul: "⥪", lhblk: "▄", LJcy: "Љ", ljcy: "љ", llarr: "⇇", ll: "≪", Ll: "⋘", llcorner: "⌞", Lleftarrow: "⇚", llhard: "⥫", lltri: "◺", Lmidot: "Ŀ", lmidot: "ŀ", lmoustache: "⎰", lmoust: "⎰", lnap: "⪉", lnapprox: "⪉", lne: "⪇", lnE: "≨", lneq: "⪇", lneqq: "≨", lnsim: "⋦", loang: "⟬", loarr: "⇽", lobrk: "⟦", longleftarrow: "⟵", LongLeftArrow: "⟵", Longleftarrow: "⟸", longleftrightarrow: "⟷", LongLeftRightArrow: "⟷", Longleftrightarrow: "⟺", longmapsto: "⟼", longrightarrow: "⟶", LongRightArrow: "⟶", Longrightarrow: "⟹", looparrowleft: "↫", looparrowright: "↬", lopar: "⦅", Lopf: "𝕃", lopf: "𝕝", loplus: "⨭", lotimes: "⨴", lowast: "∗", lowbar: "_", LowerLeftArrow: "↙", LowerRightArrow: "↘", loz: "◊", lozenge: "◊", lozf: "⧫", lpar: "(", lparlt: "⦓", lrarr: "⇆", lrcorner: "⌟", lrhar: "⇋", lrhard: "⥭", lrm: "‎", lrtri: "⊿", lsaquo: "‹", lscr: "𝓁", Lscr: "ℒ", lsh: "↰", Lsh: "↰", lsim: "≲", lsime: "⪍", lsimg: "⪏", lsqb: "[", lsquo: "‘", lsquor: "‚", Lstrok: "Ł", lstrok: "ł", ltcc: "⪦", ltcir: "⩹", lt: "<", LT: "<", Lt: "≪", ltdot: "⋖", lthree: "⋋", ltimes: "⋉", ltlarr: "⥶", ltquest: "⩻", ltri: "◃", ltrie: "⊴", ltrif: "◂", ltrPar: "⦖", lurdshar: "⥊", luruhar: "⥦", lvertneqq: "≨︀", lvnE: "≨︀", macr: "¯", male: "♂", malt: "✠", maltese: "✠", Map: "⤅", map: "↦", mapsto: "↦", mapstodown: "↧", mapstoleft: "↤", mapstoup: "↥", marker: "▮", mcomma: "⨩", Mcy: "М", mcy: "м", mdash: "—", mDDot: "∺", measuredangle: "∡", MediumSpace: " ", Mellintrf: "ℳ", Mfr: "𝔐", mfr: "𝔪", mho: "℧", micro: "µ", midast: "*", midcir: "⫰", mid: "∣", middot: "·", minusb: "⊟", minus: "−", minusd: "∸", minusdu: "⨪", MinusPlus: "∓", mlcp: "⫛", mldr: "…", mnplus: "∓", models: "⊧", Mopf: "𝕄", mopf: "𝕞", mp: "∓", mscr: "𝓂", Mscr: "ℳ", mstpos: "∾", Mu: "Μ", mu: "μ", multimap: "⊸", mumap: "⊸", nabla: "∇", Nacute: "Ń", nacute: "ń", nang: "∠⃒", nap: "≉", napE: "⩰̸", napid: "≋̸", napos: "ŉ", napprox: "≉", natural: "♮", naturals: "ℕ", natur: "♮", nbsp: " ", nbump: "≎̸", nbumpe: "≏̸", ncap: "⩃", Ncaron: "Ň", ncaron: "ň", Ncedil: "Ņ", ncedil: "ņ", ncong: "≇", ncongdot: "⩭̸", ncup: "⩂", Ncy: "Н", ncy: "н", ndash: "–", nearhk: "⤤", nearr: "↗", neArr: "⇗", nearrow: "↗", ne: "≠", nedot: "≐̸", NegativeMediumSpace: "​", NegativeThickSpace: "​", NegativeThinSpace: "​", NegativeVeryThinSpace: "​", nequiv: "≢", nesear: "⤨", nesim: "≂̸", NestedGreaterGreater: "≫", NestedLessLess: "≪", NewLine: "\n", nexist: "∄", nexists: "∄", Nfr: "𝔑", nfr: "𝔫", ngE: "≧̸", nge: "≱", ngeq: "≱", ngeqq: "≧̸", ngeqslant: "⩾̸", nges: "⩾̸", nGg: "⋙̸", ngsim: "≵", nGt: "≫⃒", ngt: "≯", ngtr: "≯", nGtv: "≫̸", nharr: "↮", nhArr: "⇎", nhpar: "⫲", ni: "∋", nis: "⋼", nisd: "⋺", niv: "∋", NJcy: "Њ", njcy: "њ", nlarr: "↚", nlArr: "⇍", nldr: "‥", nlE: "≦̸", nle: "≰", nleftarrow: "↚", nLeftarrow: "⇍", nleftrightarrow: "↮", nLeftrightarrow: "⇎", nleq: "≰", nleqq: "≦̸", nleqslant: "⩽̸", nles: "⩽̸", nless: "≮", nLl: "⋘̸", nlsim: "≴", nLt: "≪⃒", nlt: "≮", nltri: "⋪", nltrie: "⋬", nLtv: "≪̸", nmid: "∤", NoBreak: "⁠", NonBreakingSpace: " ", nopf: "𝕟", Nopf: "ℕ", Not: "⫬", not: "¬", NotCongruent: "≢", NotCupCap: "≭", NotDoubleVerticalBar: "∦", NotElement: "∉", NotEqual: "≠", NotEqualTilde: "≂̸", NotExists: "∄", NotGreater: "≯", NotGreaterEqual: "≱", NotGreaterFullEqual: "≧̸", NotGreaterGreater: "≫̸", NotGreaterLess: "≹", NotGreaterSlantEqual: "⩾̸", NotGreaterTilde: "≵", NotHumpDownHump: "≎̸", NotHumpEqual: "≏̸", notin: "∉", notindot: "⋵̸", notinE: "⋹̸", notinva: "∉", notinvb: "⋷", notinvc: "⋶", NotLeftTriangleBar: "⧏̸", NotLeftTriangle: "⋪", NotLeftTriangleEqual: "⋬", NotLess: "≮", NotLessEqual: "≰", NotLessGreater: "≸", NotLessLess: "≪̸", NotLessSlantEqual: "⩽̸", NotLessTilde: "≴", NotNestedGreaterGreater: "⪢̸", NotNestedLessLess: "⪡̸", notni: "∌", notniva: "∌", notnivb: "⋾", notnivc: "⋽", NotPrecedes: "⊀", NotPrecedesEqual: "⪯̸", NotPrecedesSlantEqual: "⋠", NotReverseElement: "∌", NotRightTriangleBar: "⧐̸", NotRightTriangle: "⋫", NotRightTriangleEqual: "⋭", NotSquareSubset: "⊏̸", NotSquareSubsetEqual: "⋢", NotSquareSuperset: "⊐̸", NotSquareSupersetEqual: "⋣", NotSubset: "⊂⃒", NotSubsetEqual: "⊈", NotSucceeds: "⊁", NotSucceedsEqual: "⪰̸", NotSucceedsSlantEqual: "⋡", NotSucceedsTilde: "≿̸", NotSuperset: "⊃⃒", NotSupersetEqual: "⊉", NotTilde: "≁", NotTildeEqual: "≄", NotTildeFullEqual: "≇", NotTildeTilde: "≉", NotVerticalBar: "∤", nparallel: "∦", npar: "∦", nparsl: "⫽⃥", npart: "∂̸", npolint: "⨔", npr: "⊀", nprcue: "⋠", nprec: "⊀", npreceq: "⪯̸", npre: "⪯̸", nrarrc: "⤳̸", nrarr: "↛", nrArr: "⇏", nrarrw: "↝̸", nrightarrow: "↛", nRightarrow: "⇏", nrtri: "⋫", nrtrie: "⋭", nsc: "⊁", nsccue: "⋡", nsce: "⪰̸", Nscr: "𝒩", nscr: "𝓃", nshortmid: "∤", nshortparallel: "∦", nsim: "≁", nsime: "≄", nsimeq: "≄", nsmid: "∤", nspar: "∦", nsqsube: "⋢", nsqsupe: "⋣", nsub: "⊄", nsubE: "⫅̸", nsube: "⊈", nsubset: "⊂⃒", nsubseteq: "⊈", nsubseteqq: "⫅̸", nsucc: "⊁", nsucceq: "⪰̸", nsup: "⊅", nsupE: "⫆̸", nsupe: "⊉", nsupset: "⊃⃒", nsupseteq: "⊉", nsupseteqq: "⫆̸", ntgl: "≹", Ntilde: "Ñ", ntilde: "ñ", ntlg: "≸", ntriangleleft: "⋪", ntrianglelefteq: "⋬", ntriangleright: "⋫", ntrianglerighteq: "⋭", Nu: "Ν", nu: "ν", num: "#", numero: "№", numsp: " ", nvap: "≍⃒", nvdash: "⊬", nvDash: "⊭", nVdash: "⊮", nVDash: "⊯", nvge: "≥⃒", nvgt: ">⃒", nvHarr: "⤄", nvinfin: "⧞", nvlArr: "⤂", nvle: "≤⃒", nvlt: "<⃒", nvltrie: "⊴⃒", nvrArr: "⤃", nvrtrie: "⊵⃒", nvsim: "∼⃒", nwarhk: "⤣", nwarr: "↖", nwArr: "⇖", nwarrow: "↖", nwnear: "⤧", Oacute: "Ó", oacute: "ó", oast: "⊛", Ocirc: "Ô", ocirc: "ô", ocir: "⊚", Ocy: "О", ocy: "о", odash: "⊝", Odblac: "Ő", odblac: "ő", odiv: "⨸", odot: "⊙", odsold: "⦼", OElig: "Œ", oelig: "œ", ofcir: "⦿", Ofr: "𝔒", ofr: "𝔬", ogon: "˛", Ograve: "Ò", ograve: "ò", ogt: "⧁", ohbar: "⦵", ohm: "Ω", oint: "∮", olarr: "↺", olcir: "⦾", olcross: "⦻", oline: "‾", olt: "⧀", Omacr: "Ō", omacr: "ō", Omega: "Ω", omega: "ω", Omicron: "Ο", omicron: "ο", omid: "⦶", ominus: "⊖", Oopf: "𝕆", oopf: "𝕠", opar: "⦷", OpenCurlyDoubleQuote: "“", OpenCurlyQuote: "‘", operp: "⦹", oplus: "⊕", orarr: "↻", Or: "⩔", or: "∨", ord: "⩝", order: "ℴ", orderof: "ℴ", ordf: "ª", ordm: "º", origof: "⊶", oror: "⩖", orslope: "⩗", orv: "⩛", oS: "Ⓢ", Oscr: "𝒪", oscr: "ℴ", Oslash: "Ø", oslash: "ø", osol: "⊘", Otilde: "Õ", otilde: "õ", otimesas: "⨶", Otimes: "⨷", otimes: "⊗", Ouml: "Ö", ouml: "ö", ovbar: "⌽", OverBar: "‾", OverBrace: "⏞", OverBracket: "⎴", OverParenthesis: "⏜", para: "¶", parallel: "∥", par: "∥", parsim: "⫳", parsl: "⫽", part: "∂", PartialD: "∂", Pcy: "П", pcy: "п", percnt: "%", period: ".", permil: "‰", perp: "⊥", pertenk: "‱", Pfr: "𝔓", pfr: "𝔭", Phi: "Φ", phi: "φ", phiv: "ϕ", phmmat: "ℳ", phone: "☎", Pi: "Π", pi: "π", pitchfork: "⋔", piv: "ϖ", planck: "ℏ", planckh: "ℎ", plankv: "ℏ", plusacir: "⨣", plusb: "⊞", pluscir: "⨢", plus: "+", plusdo: "∔", plusdu: "⨥", pluse: "⩲", PlusMinus: "±", plusmn: "±", plussim: "⨦", plustwo: "⨧", pm: "±", Poincareplane: "ℌ", pointint: "⨕", popf: "𝕡", Popf: "ℙ", pound: "£", prap: "⪷", Pr: "⪻", pr: "≺", prcue: "≼", precapprox: "⪷", prec: "≺", preccurlyeq: "≼", Precedes: "≺", PrecedesEqual: "⪯", PrecedesSlantEqual: "≼", PrecedesTilde: "≾", preceq: "⪯", precnapprox: "⪹", precneqq: "⪵", precnsim: "⋨", pre: "⪯", prE: "⪳", precsim: "≾", prime: "′", Prime: "″", primes: "ℙ", prnap: "⪹", prnE: "⪵", prnsim: "⋨", prod: "∏", Product: "∏", profalar: "⌮", profline: "⌒", profsurf: "⌓", prop: "∝", Proportional: "∝", Proportion: "∷", propto: "∝", prsim: "≾", prurel: "⊰", Pscr: "𝒫", pscr: "𝓅", Psi: "Ψ", psi: "ψ", puncsp: " ", Qfr: "𝔔", qfr: "𝔮", qint: "⨌", qopf: "𝕢", Qopf: "ℚ", qprime: "⁗", Qscr: "𝒬", qscr: "𝓆", quaternions: "ℍ", quatint: "⨖", quest: "?", questeq: "≟", quot: '"', QUOT: '"', rAarr: "⇛", race: "∽̱", Racute: "Ŕ", racute: "ŕ", radic: "√", raemptyv: "⦳", rang: "⟩", Rang: "⟫", rangd: "⦒", range: "⦥", rangle: "⟩", raquo: "»", rarrap: "⥵", rarrb: "⇥", rarrbfs: "⤠", rarrc: "⤳", rarr: "→", Rarr: "↠", rArr: "⇒", rarrfs: "⤞", rarrhk: "↪", rarrlp: "↬", rarrpl: "⥅", rarrsim: "⥴", Rarrtl: "⤖", rarrtl: "↣", rarrw: "↝", ratail: "⤚", rAtail: "⤜", ratio: "∶", rationals: "ℚ", rbarr: "⤍", rBarr: "⤏", RBarr: "⤐", rbbrk: "❳", rbrace: "}", rbrack: "]", rbrke: "⦌", rbrksld: "⦎", rbrkslu: "⦐", Rcaron: "Ř", rcaron: "ř", Rcedil: "Ŗ", rcedil: "ŗ", rceil: "⌉", rcub: "}", Rcy: "Р", rcy: "р", rdca: "⤷", rdldhar: "⥩", rdquo: "”", rdquor: "”", rdsh: "↳", real: "ℜ", realine: "ℛ", realpart: "ℜ", reals: "ℝ", Re: "ℜ", rect: "▭", reg: "®", REG: "®", ReverseElement: "∋", ReverseEquilibrium: "⇋", ReverseUpEquilibrium: "⥯", rfisht: "⥽", rfloor: "⌋", rfr: "𝔯", Rfr: "ℜ", rHar: "⥤", rhard: "⇁", rharu: "⇀", rharul: "⥬", Rho: "Ρ", rho: "ρ", rhov: "ϱ", RightAngleBracket: "⟩", RightArrowBar: "⇥", rightarrow: "→", RightArrow: "→", Rightarrow: "⇒", RightArrowLeftArrow: "⇄", rightarrowtail: "↣", RightCeiling: "⌉", RightDoubleBracket: "⟧", RightDownTeeVector: "⥝", RightDownVectorBar: "⥕", RightDownVector: "⇂", RightFloor: "⌋", rightharpoondown: "⇁", rightharpoonup: "⇀", rightleftarrows: "⇄", rightleftharpoons: "⇌", rightrightarrows: "⇉", rightsquigarrow: "↝", RightTeeArrow: "↦", RightTee: "⊢", RightTeeVector: "⥛", rightthreetimes: "⋌", RightTriangleBar: "⧐", RightTriangle: "⊳", RightTriangleEqual: "⊵", RightUpDownVector: "⥏", RightUpTeeVector: "⥜", RightUpVectorBar: "⥔", RightUpVector: "↾", RightVectorBar: "⥓", RightVector: "⇀", ring: "˚", risingdotseq: "≓", rlarr: "⇄", rlhar: "⇌", rlm: "‏", rmoustache: "⎱", rmoust: "⎱", rnmid: "⫮", roang: "⟭", roarr: "⇾", robrk: "⟧", ropar: "⦆", ropf: "𝕣", Ropf: "ℝ", roplus: "⨮", rotimes: "⨵", RoundImplies: "⥰", rpar: ")", rpargt: "⦔", rppolint: "⨒", rrarr: "⇉", Rrightarrow: "⇛", rsaquo: "›", rscr: "𝓇", Rscr: "ℛ", rsh: "↱", Rsh: "↱", rsqb: "]", rsquo: "’", rsquor: "’", rthree: "⋌", rtimes: "⋊", rtri: "▹", rtrie: "⊵", rtrif: "▸", rtriltri: "⧎", RuleDelayed: "⧴", ruluhar: "⥨", rx: "℞", Sacute: "Ś", sacute: "ś", sbquo: "‚", scap: "⪸", Scaron: "Š", scaron: "š", Sc: "⪼", sc: "≻", sccue: "≽", sce: "⪰", scE: "⪴", Scedil: "Ş", scedil: "ş", Scirc: "Ŝ", scirc: "ŝ", scnap: "⪺", scnE: "⪶", scnsim: "⋩", scpolint: "⨓", scsim: "≿", Scy: "С", scy: "с", sdotb: "⊡", sdot: "⋅", sdote: "⩦", searhk: "⤥", searr: "↘", seArr: "⇘", searrow: "↘", sect: "§", semi: ";", seswar: "⤩", setminus: "∖", setmn: "∖", sext: "✶", Sfr: "𝔖", sfr: "𝔰", sfrown: "⌢", sharp: "♯", SHCHcy: "Щ", shchcy: "щ", SHcy: "Ш", shcy: "ш", ShortDownArrow: "↓", ShortLeftArrow: "←", shortmid: "∣", shortparallel: "∥", ShortRightArrow: "→", ShortUpArrow: "↑", shy: "­", Sigma: "Σ", sigma: "σ", sigmaf: "ς", sigmav: "ς", sim: "∼", simdot: "⩪", sime: "≃", simeq: "≃", simg: "⪞", simgE: "⪠", siml: "⪝", simlE: "⪟", simne: "≆", simplus: "⨤", simrarr: "⥲", slarr: "←", SmallCircle: "∘", smallsetminus: "∖", smashp: "⨳", smeparsl: "⧤", smid: "∣", smile: "⌣", smt: "⪪", smte: "⪬", smtes: "⪬︀", SOFTcy: "Ь", softcy: "ь", solbar: "⌿", solb: "⧄", sol: "/", Sopf: "𝕊", sopf: "𝕤", spades: "♠", spadesuit: "♠", spar: "∥", sqcap: "⊓", sqcaps: "⊓︀", sqcup: "⊔", sqcups: "⊔︀", Sqrt: "√", sqsub: "⊏", sqsube: "⊑", sqsubset: "⊏", sqsubseteq: "⊑", sqsup: "⊐", sqsupe: "⊒", sqsupset: "⊐", sqsupseteq: "⊒", square: "□", Square: "□", SquareIntersection: "⊓", SquareSubset: "⊏", SquareSubsetEqual: "⊑", SquareSuperset: "⊐", SquareSupersetEqual: "⊒", SquareUnion: "⊔", squarf: "▪", squ: "□", squf: "▪", srarr: "→", Sscr: "𝒮", sscr: "𝓈", ssetmn: "∖", ssmile: "⌣", sstarf: "⋆", Star: "⋆", star: "☆", starf: "★", straightepsilon: "ϵ", straightphi: "ϕ", strns: "¯", sub: "⊂", Sub: "⋐", subdot: "⪽", subE: "⫅", sube: "⊆", subedot: "⫃", submult: "⫁", subnE: "⫋", subne: "⊊", subplus: "⪿", subrarr: "⥹", subset: "⊂", Subset: "⋐", subseteq: "⊆", subseteqq: "⫅", SubsetEqual: "⊆", subsetneq: "⊊", subsetneqq: "⫋", subsim: "⫇", subsub: "⫕", subsup: "⫓", succapprox: "⪸", succ: "≻", succcurlyeq: "≽", Succeeds: "≻", SucceedsEqual: "⪰", SucceedsSlantEqual: "≽", SucceedsTilde: "≿", succeq: "⪰", succnapprox: "⪺", succneqq: "⪶", succnsim: "⋩", succsim: "≿", SuchThat: "∋", sum: "∑", Sum: "∑", sung: "♪", sup1: "¹", sup2: "²", sup3: "³", sup: "⊃", Sup: "⋑", supdot: "⪾", supdsub: "⫘", supE: "⫆", supe: "⊇", supedot: "⫄", Superset: "⊃", SupersetEqual: "⊇", suphsol: "⟉", suphsub: "⫗", suplarr: "⥻", supmult: "⫂", supnE: "⫌", supne: "⊋", supplus: "⫀", supset: "⊃", Supset: "⋑", supseteq: "⊇", supseteqq: "⫆", supsetneq: "⊋", supsetneqq: "⫌", supsim: "⫈", supsub: "⫔", supsup: "⫖", swarhk: "⤦", swarr: "↙", swArr: "⇙", swarrow: "↙", swnwar: "⤪", szlig: "ß", Tab: "	", target: "⌖", Tau: "Τ", tau: "τ", tbrk: "⎴", Tcaron: "Ť", tcaron: "ť", Tcedil: "Ţ", tcedil: "ţ", Tcy: "Т", tcy: "т", tdot: "⃛", telrec: "⌕", Tfr: "𝔗", tfr: "𝔱", there4: "∴", therefore: "∴", Therefore: "∴", Theta: "Θ", theta: "θ", thetasym: "ϑ", thetav: "ϑ", thickapprox: "≈", thicksim: "∼", ThickSpace: "  ", ThinSpace: " ", thinsp: " ", thkap: "≈", thksim: "∼", THORN: "Þ", thorn: "þ", tilde: "˜", Tilde: "∼", TildeEqual: "≃", TildeFullEqual: "≅", TildeTilde: "≈", timesbar: "⨱", timesb: "⊠", times: "×", timesd: "⨰", tint: "∭", toea: "⤨", topbot: "⌶", topcir: "⫱", top: "⊤", Topf: "𝕋", topf: "𝕥", topfork: "⫚", tosa: "⤩", tprime: "‴", trade: "™", TRADE: "™", triangle: "▵", triangledown: "▿", triangleleft: "◃", trianglelefteq: "⊴", triangleq: "≜", triangleright: "▹", trianglerighteq: "⊵", tridot: "◬", trie: "≜", triminus: "⨺", TripleDot: "⃛", triplus: "⨹", trisb: "⧍", tritime: "⨻", trpezium: "⏢", Tscr: "𝒯", tscr: "𝓉", TScy: "Ц", tscy: "ц", TSHcy: "Ћ", tshcy: "ћ", Tstrok: "Ŧ", tstrok: "ŧ", twixt: "≬", twoheadleftarrow: "↞", twoheadrightarrow: "↠", Uacute: "Ú", uacute: "ú", uarr: "↑", Uarr: "↟", uArr: "⇑", Uarrocir: "⥉", Ubrcy: "Ў", ubrcy: "ў", Ubreve: "Ŭ", ubreve: "ŭ", Ucirc: "Û", ucirc: "û", Ucy: "У", ucy: "у", udarr: "⇅", Udblac: "Ű", udblac: "ű", udhar: "⥮", ufisht: "⥾", Ufr: "𝔘", ufr: "𝔲", Ugrave: "Ù", ugrave: "ù", uHar: "⥣", uharl: "↿", uharr: "↾", uhblk: "▀", ulcorn: "⌜", ulcorner: "⌜", ulcrop: "⌏", ultri: "◸", Umacr: "Ū", umacr: "ū", uml: "¨", UnderBar: "_", UnderBrace: "⏟", UnderBracket: "⎵", UnderParenthesis: "⏝", Union: "⋃", UnionPlus: "⊎", Uogon: "Ų", uogon: "ų", Uopf: "𝕌", uopf: "𝕦", UpArrowBar: "⤒", uparrow: "↑", UpArrow: "↑", Uparrow: "⇑", UpArrowDownArrow: "⇅", updownarrow: "↕", UpDownArrow: "↕", Updownarrow: "⇕", UpEquilibrium: "⥮", upharpoonleft: "↿", upharpoonright: "↾", uplus: "⊎", UpperLeftArrow: "↖", UpperRightArrow: "↗", upsi: "υ", Upsi: "ϒ", upsih: "ϒ", Upsilon: "Υ", upsilon: "υ", UpTeeArrow: "↥", UpTee: "⊥", upuparrows: "⇈", urcorn: "⌝", urcorner: "⌝", urcrop: "⌎", Uring: "Ů", uring: "ů", urtri: "◹", Uscr: "𝒰", uscr: "𝓊", utdot: "⋰", Utilde: "Ũ", utilde: "ũ", utri: "▵", utrif: "▴", uuarr: "⇈", Uuml: "Ü", uuml: "ü", uwangle: "⦧", vangrt: "⦜", varepsilon: "ϵ", varkappa: "ϰ", varnothing: "∅", varphi: "ϕ", varpi: "ϖ", varpropto: "∝", varr: "↕", vArr: "⇕", varrho: "ϱ", varsigma: "ς", varsubsetneq: "⊊︀", varsubsetneqq: "⫋︀", varsupsetneq: "⊋︀", varsupsetneqq: "⫌︀", vartheta: "ϑ", vartriangleleft: "⊲", vartriangleright: "⊳", vBar: "⫨", Vbar: "⫫", vBarv: "⫩", Vcy: "В", vcy: "в", vdash: "⊢", vDash: "⊨", Vdash: "⊩", VDash: "⊫", Vdashl: "⫦", veebar: "⊻", vee: "∨", Vee: "⋁", veeeq: "≚", vellip: "⋮", verbar: "|", Verbar: "‖", vert: "|", Vert: "‖", VerticalBar: "∣", VerticalLine: "|", VerticalSeparator: "❘", VerticalTilde: "≀", VeryThinSpace: " ", Vfr: "𝔙", vfr: "𝔳", vltri: "⊲", vnsub: "⊂⃒", vnsup: "⊃⃒", Vopf: "𝕍", vopf: "𝕧", vprop: "∝", vrtri: "⊳", Vscr: "𝒱", vscr: "𝓋", vsubnE: "⫋︀", vsubne: "⊊︀", vsupnE: "⫌︀", vsupne: "⊋︀", Vvdash: "⊪", vzigzag: "⦚", Wcirc: "Ŵ", wcirc: "ŵ", wedbar: "⩟", wedge: "∧", Wedge: "⋀", wedgeq: "≙", weierp: "℘", Wfr: "𝔚", wfr: "𝔴", Wopf: "𝕎", wopf: "𝕨", wp: "℘", wr: "≀", wreath: "≀", Wscr: "𝒲", wscr: "𝓌", xcap: "⋂", xcirc: "◯", xcup: "⋃", xdtri: "▽", Xfr: "𝔛", xfr: "𝔵", xharr: "⟷", xhArr: "⟺", Xi: "Ξ", xi: "ξ", xlarr: "⟵", xlArr: "⟸", xmap: "⟼", xnis: "⋻", xodot: "⨀", Xopf: "𝕏", xopf: "𝕩", xoplus: "⨁", xotime: "⨂", xrarr: "⟶", xrArr: "⟹", Xscr: "𝒳", xscr: "𝓍", xsqcup: "⨆", xuplus: "⨄", xutri: "△", xvee: "⋁", xwedge: "⋀", Yacute: "Ý", yacute: "ý", YAcy: "Я", yacy: "я", Ycirc: "Ŷ", ycirc: "ŷ", Ycy: "Ы", ycy: "ы", yen: "¥", Yfr: "𝔜", yfr: "𝔶", YIcy: "Ї", yicy: "ї", Yopf: "𝕐", yopf: "𝕪", Yscr: "𝒴", yscr: "𝓎", YUcy: "Ю", yucy: "ю", yuml: "ÿ", Yuml: "Ÿ", Zacute: "Ź", zacute: "ź", Zcaron: "Ž", zcaron: "ž", Zcy: "З", zcy: "з", Zdot: "Ż", zdot: "ż", zeetrf: "ℨ", ZeroWidthSpace: "​", Zeta: "Ζ", zeta: "ζ", zfr: "𝔷", Zfr: "ℨ", ZHcy: "Ж", zhcy: "ж", zigrarr: "⇝", zopf: "𝕫", Zopf: "ℤ", Zscr: "𝒵", zscr: "𝓏", zwj: "‍", zwnj: "‌" };
  }
});

// node_modules/entities/maps/legacy.json
var require_legacy = __commonJS({
  "node_modules/entities/maps/legacy.json"(exports2, module2) {
    module2.exports = { Aacute: "Á", aacute: "á", Acirc: "Â", acirc: "â", acute: "´", AElig: "Æ", aelig: "æ", Agrave: "À", agrave: "à", amp: "&", AMP: "&", Aring: "Å", aring: "å", Atilde: "Ã", atilde: "ã", Auml: "Ä", auml: "ä", brvbar: "¦", Ccedil: "Ç", ccedil: "ç", cedil: "¸", cent: "¢", copy: "©", COPY: "©", curren: "¤", deg: "°", divide: "÷", Eacute: "É", eacute: "é", Ecirc: "Ê", ecirc: "ê", Egrave: "È", egrave: "è", ETH: "Ð", eth: "ð", Euml: "Ë", euml: "ë", frac12: "½", frac14: "¼", frac34: "¾", gt: ">", GT: ">", Iacute: "Í", iacute: "í", Icirc: "Î", icirc: "î", iexcl: "¡", Igrave: "Ì", igrave: "ì", iquest: "¿", Iuml: "Ï", iuml: "ï", laquo: "«", lt: "<", LT: "<", macr: "¯", micro: "µ", middot: "·", nbsp: " ", not: "¬", Ntilde: "Ñ", ntilde: "ñ", Oacute: "Ó", oacute: "ó", Ocirc: "Ô", ocirc: "ô", Ograve: "Ò", ograve: "ò", ordf: "ª", ordm: "º", Oslash: "Ø", oslash: "ø", Otilde: "Õ", otilde: "õ", Ouml: "Ö", ouml: "ö", para: "¶", plusmn: "±", pound: "£", quot: '"', QUOT: '"', raquo: "»", reg: "®", REG: "®", sect: "§", shy: "­", sup1: "¹", sup2: "²", sup3: "³", szlig: "ß", THORN: "Þ", thorn: "þ", times: "×", Uacute: "Ú", uacute: "ú", Ucirc: "Û", ucirc: "û", Ugrave: "Ù", ugrave: "ù", uml: "¨", Uuml: "Ü", uuml: "ü", Yacute: "Ý", yacute: "ý", yen: "¥", yuml: "ÿ" };
  }
});

// node_modules/entities/maps/xml.json
var require_xml = __commonJS({
  "node_modules/entities/maps/xml.json"(exports2, module2) {
    module2.exports = { amp: "&", apos: "'", gt: ">", lt: "<", quot: '"' };
  }
});

// node_modules/htmlparser2/lib/Tokenizer.js
var require_Tokenizer = __commonJS({
  "node_modules/htmlparser2/lib/Tokenizer.js"(exports2, module2) {
    init_process_shim();
    module2.exports = Tokenizer;
    var decodeCodePoint = require_decode_codepoint();
    var entityMap = require_entities();
    var legacyMap = require_legacy();
    var xmlMap = require_xml();
    var i = 0;
    var TEXT = i++;
    var BEFORE_TAG_NAME = i++;
    var IN_TAG_NAME = i++;
    var IN_SELF_CLOSING_TAG = i++;
    var BEFORE_CLOSING_TAG_NAME = i++;
    var IN_CLOSING_TAG_NAME = i++;
    var AFTER_CLOSING_TAG_NAME = i++;
    var BEFORE_ATTRIBUTE_NAME = i++;
    var IN_ATTRIBUTE_NAME = i++;
    var AFTER_ATTRIBUTE_NAME = i++;
    var BEFORE_ATTRIBUTE_VALUE = i++;
    var IN_ATTRIBUTE_VALUE_DQ = i++;
    var IN_ATTRIBUTE_VALUE_SQ = i++;
    var IN_ATTRIBUTE_VALUE_NQ = i++;
    var BEFORE_DECLARATION = i++;
    var IN_DECLARATION = i++;
    var IN_PROCESSING_INSTRUCTION = i++;
    var BEFORE_COMMENT = i++;
    var IN_COMMENT = i++;
    var AFTER_COMMENT_1 = i++;
    var AFTER_COMMENT_2 = i++;
    var BEFORE_CDATA_1 = i++;
    var BEFORE_CDATA_2 = i++;
    var BEFORE_CDATA_3 = i++;
    var BEFORE_CDATA_4 = i++;
    var BEFORE_CDATA_5 = i++;
    var BEFORE_CDATA_6 = i++;
    var IN_CDATA = i++;
    var AFTER_CDATA_1 = i++;
    var AFTER_CDATA_2 = i++;
    var BEFORE_SPECIAL = i++;
    var BEFORE_SPECIAL_END = i++;
    var BEFORE_SCRIPT_1 = i++;
    var BEFORE_SCRIPT_2 = i++;
    var BEFORE_SCRIPT_3 = i++;
    var BEFORE_SCRIPT_4 = i++;
    var BEFORE_SCRIPT_5 = i++;
    var AFTER_SCRIPT_1 = i++;
    var AFTER_SCRIPT_2 = i++;
    var AFTER_SCRIPT_3 = i++;
    var AFTER_SCRIPT_4 = i++;
    var AFTER_SCRIPT_5 = i++;
    var BEFORE_STYLE_1 = i++;
    var BEFORE_STYLE_2 = i++;
    var BEFORE_STYLE_3 = i++;
    var BEFORE_STYLE_4 = i++;
    var AFTER_STYLE_1 = i++;
    var AFTER_STYLE_2 = i++;
    var AFTER_STYLE_3 = i++;
    var AFTER_STYLE_4 = i++;
    var BEFORE_ENTITY = i++;
    var BEFORE_NUMERIC_ENTITY = i++;
    var IN_NAMED_ENTITY = i++;
    var IN_NUMERIC_ENTITY = i++;
    var IN_HEX_ENTITY = i++;
    var j = 0;
    var SPECIAL_NONE = j++;
    var SPECIAL_SCRIPT = j++;
    var SPECIAL_STYLE = j++;
    function whitespace(c) {
      return c === " " || c === "\n" || c === "	" || c === "\f" || c === "\r";
    }
    function ifElseState(upper, SUCCESS, FAILURE) {
      var lower = upper.toLowerCase();
      if (upper === lower) {
        return function(c) {
          if (c === lower) {
            this._state = SUCCESS;
          } else {
            this._state = FAILURE;
            this._index--;
          }
        };
      } else {
        return function(c) {
          if (c === lower || c === upper) {
            this._state = SUCCESS;
          } else {
            this._state = FAILURE;
            this._index--;
          }
        };
      }
    }
    function consumeSpecialNameChar(upper, NEXT_STATE) {
      var lower = upper.toLowerCase();
      return function(c) {
        if (c === lower || c === upper) {
          this._state = NEXT_STATE;
        } else {
          this._state = IN_TAG_NAME;
          this._index--;
        }
      };
    }
    function Tokenizer(options, cbs) {
      this._state = TEXT;
      this._buffer = "";
      this._sectionStart = 0;
      this._index = 0;
      this._bufferOffset = 0;
      this._baseState = TEXT;
      this._special = SPECIAL_NONE;
      this._cbs = cbs;
      this._running = true;
      this._ended = false;
      this._xmlMode = !!(options && options.xmlMode);
      this._decodeEntities = !!(options && options.decodeEntities);
    }
    Tokenizer.prototype._stateText = function(c) {
      if (c === "<") {
        if (this._index > this._sectionStart) {
          this._cbs.ontext(this._getSection());
        }
        this._state = BEFORE_TAG_NAME;
        this._sectionStart = this._index;
      } else if (this._decodeEntities && this._special === SPECIAL_NONE && c === "&") {
        if (this._index > this._sectionStart) {
          this._cbs.ontext(this._getSection());
        }
        this._baseState = TEXT;
        this._state = BEFORE_ENTITY;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateBeforeTagName = function(c) {
      if (c === "/") {
        this._state = BEFORE_CLOSING_TAG_NAME;
      } else if (c === "<") {
        this._cbs.ontext(this._getSection());
        this._sectionStart = this._index;
      } else if (c === ">" || this._special !== SPECIAL_NONE || whitespace(c)) {
        this._state = TEXT;
      } else if (c === "!") {
        this._state = BEFORE_DECLARATION;
        this._sectionStart = this._index + 1;
      } else if (c === "?") {
        this._state = IN_PROCESSING_INSTRUCTION;
        this._sectionStart = this._index + 1;
      } else {
        this._state = !this._xmlMode && (c === "s" || c === "S") ? BEFORE_SPECIAL : IN_TAG_NAME;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateInTagName = function(c) {
      if (c === "/" || c === ">" || whitespace(c)) {
        this._emitToken("onopentagname");
        this._state = BEFORE_ATTRIBUTE_NAME;
        this._index--;
      }
    };
    Tokenizer.prototype._stateBeforeCloseingTagName = function(c) {
      if (whitespace(c)) ;
      else if (c === ">") {
        this._state = TEXT;
      } else if (this._special !== SPECIAL_NONE) {
        if (c === "s" || c === "S") {
          this._state = BEFORE_SPECIAL_END;
        } else {
          this._state = TEXT;
          this._index--;
        }
      } else {
        this._state = IN_CLOSING_TAG_NAME;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateInCloseingTagName = function(c) {
      if (c === ">" || whitespace(c)) {
        this._emitToken("onclosetag");
        this._state = AFTER_CLOSING_TAG_NAME;
        this._index--;
      }
    };
    Tokenizer.prototype._stateAfterCloseingTagName = function(c) {
      if (c === ">") {
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      }
    };
    Tokenizer.prototype._stateBeforeAttributeName = function(c) {
      if (c === ">") {
        this._cbs.onopentagend();
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      } else if (c === "/") {
        this._state = IN_SELF_CLOSING_TAG;
      } else if (!whitespace(c)) {
        this._state = IN_ATTRIBUTE_NAME;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateInSelfClosingTag = function(c) {
      if (c === ">") {
        this._cbs.onselfclosingtag();
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      } else if (!whitespace(c)) {
        this._state = BEFORE_ATTRIBUTE_NAME;
        this._index--;
      }
    };
    Tokenizer.prototype._stateInAttributeName = function(c) {
      if (c === "=" || c === "/" || c === ">" || whitespace(c)) {
        this._cbs.onattribname(this._getSection());
        this._sectionStart = -1;
        this._state = AFTER_ATTRIBUTE_NAME;
        this._index--;
      }
    };
    Tokenizer.prototype._stateAfterAttributeName = function(c) {
      if (c === "=") {
        this._state = BEFORE_ATTRIBUTE_VALUE;
      } else if (c === "/" || c === ">") {
        this._cbs.onattribend();
        this._state = BEFORE_ATTRIBUTE_NAME;
        this._index--;
      } else if (!whitespace(c)) {
        this._cbs.onattribend();
        this._state = IN_ATTRIBUTE_NAME;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateBeforeAttributeValue = function(c) {
      if (c === '"') {
        this._state = IN_ATTRIBUTE_VALUE_DQ;
        this._sectionStart = this._index + 1;
      } else if (c === "'") {
        this._state = IN_ATTRIBUTE_VALUE_SQ;
        this._sectionStart = this._index + 1;
      } else if (!whitespace(c)) {
        this._state = IN_ATTRIBUTE_VALUE_NQ;
        this._sectionStart = this._index;
        this._index--;
      }
    };
    Tokenizer.prototype._stateInAttributeValueDoubleQuotes = function(c) {
      if (c === '"') {
        this._emitToken("onattribdata");
        this._cbs.onattribend();
        this._state = BEFORE_ATTRIBUTE_NAME;
      } else if (this._decodeEntities && c === "&") {
        this._emitToken("onattribdata");
        this._baseState = this._state;
        this._state = BEFORE_ENTITY;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateInAttributeValueSingleQuotes = function(c) {
      if (c === "'") {
        this._emitToken("onattribdata");
        this._cbs.onattribend();
        this._state = BEFORE_ATTRIBUTE_NAME;
      } else if (this._decodeEntities && c === "&") {
        this._emitToken("onattribdata");
        this._baseState = this._state;
        this._state = BEFORE_ENTITY;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateInAttributeValueNoQuotes = function(c) {
      if (whitespace(c) || c === ">") {
        this._emitToken("onattribdata");
        this._cbs.onattribend();
        this._state = BEFORE_ATTRIBUTE_NAME;
        this._index--;
      } else if (this._decodeEntities && c === "&") {
        this._emitToken("onattribdata");
        this._baseState = this._state;
        this._state = BEFORE_ENTITY;
        this._sectionStart = this._index;
      }
    };
    Tokenizer.prototype._stateBeforeDeclaration = function(c) {
      this._state = c === "[" ? BEFORE_CDATA_1 : c === "-" ? BEFORE_COMMENT : IN_DECLARATION;
    };
    Tokenizer.prototype._stateInDeclaration = function(c) {
      if (c === ">") {
        this._cbs.ondeclaration(this._getSection());
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      }
    };
    Tokenizer.prototype._stateInProcessingInstruction = function(c) {
      if (c === ">") {
        this._cbs.onprocessinginstruction(this._getSection());
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      }
    };
    Tokenizer.prototype._stateBeforeComment = function(c) {
      if (c === "-") {
        this._state = IN_COMMENT;
        this._sectionStart = this._index + 1;
      } else {
        this._state = IN_DECLARATION;
      }
    };
    Tokenizer.prototype._stateInComment = function(c) {
      if (c === "-") this._state = AFTER_COMMENT_1;
    };
    Tokenizer.prototype._stateAfterComment1 = function(c) {
      if (c === "-") {
        this._state = AFTER_COMMENT_2;
      } else {
        this._state = IN_COMMENT;
      }
    };
    Tokenizer.prototype._stateAfterComment2 = function(c) {
      if (c === ">") {
        this._cbs.oncomment(
          this._buffer.substring(this._sectionStart, this._index - 2)
        );
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      } else if (c !== "-") {
        this._state = IN_COMMENT;
      }
    };
    Tokenizer.prototype._stateBeforeCdata1 = ifElseState(
      "C",
      BEFORE_CDATA_2,
      IN_DECLARATION
    );
    Tokenizer.prototype._stateBeforeCdata2 = ifElseState(
      "D",
      BEFORE_CDATA_3,
      IN_DECLARATION
    );
    Tokenizer.prototype._stateBeforeCdata3 = ifElseState(
      "A",
      BEFORE_CDATA_4,
      IN_DECLARATION
    );
    Tokenizer.prototype._stateBeforeCdata4 = ifElseState(
      "T",
      BEFORE_CDATA_5,
      IN_DECLARATION
    );
    Tokenizer.prototype._stateBeforeCdata5 = ifElseState(
      "A",
      BEFORE_CDATA_6,
      IN_DECLARATION
    );
    Tokenizer.prototype._stateBeforeCdata6 = function(c) {
      if (c === "[") {
        this._state = IN_CDATA;
        this._sectionStart = this._index + 1;
      } else {
        this._state = IN_DECLARATION;
        this._index--;
      }
    };
    Tokenizer.prototype._stateInCdata = function(c) {
      if (c === "]") this._state = AFTER_CDATA_1;
    };
    Tokenizer.prototype._stateAfterCdata1 = function(c) {
      if (c === "]") this._state = AFTER_CDATA_2;
      else this._state = IN_CDATA;
    };
    Tokenizer.prototype._stateAfterCdata2 = function(c) {
      if (c === ">") {
        this._cbs.oncdata(
          this._buffer.substring(this._sectionStart, this._index - 2)
        );
        this._state = TEXT;
        this._sectionStart = this._index + 1;
      } else if (c !== "]") {
        this._state = IN_CDATA;
      }
    };
    Tokenizer.prototype._stateBeforeSpecial = function(c) {
      if (c === "c" || c === "C") {
        this._state = BEFORE_SCRIPT_1;
      } else if (c === "t" || c === "T") {
        this._state = BEFORE_STYLE_1;
      } else {
        this._state = IN_TAG_NAME;
        this._index--;
      }
    };
    Tokenizer.prototype._stateBeforeSpecialEnd = function(c) {
      if (this._special === SPECIAL_SCRIPT && (c === "c" || c === "C")) {
        this._state = AFTER_SCRIPT_1;
      } else if (this._special === SPECIAL_STYLE && (c === "t" || c === "T")) {
        this._state = AFTER_STYLE_1;
      } else this._state = TEXT;
    };
    Tokenizer.prototype._stateBeforeScript1 = consumeSpecialNameChar(
      "R",
      BEFORE_SCRIPT_2
    );
    Tokenizer.prototype._stateBeforeScript2 = consumeSpecialNameChar(
      "I",
      BEFORE_SCRIPT_3
    );
    Tokenizer.prototype._stateBeforeScript3 = consumeSpecialNameChar(
      "P",
      BEFORE_SCRIPT_4
    );
    Tokenizer.prototype._stateBeforeScript4 = consumeSpecialNameChar(
      "T",
      BEFORE_SCRIPT_5
    );
    Tokenizer.prototype._stateBeforeScript5 = function(c) {
      if (c === "/" || c === ">" || whitespace(c)) {
        this._special = SPECIAL_SCRIPT;
      }
      this._state = IN_TAG_NAME;
      this._index--;
    };
    Tokenizer.prototype._stateAfterScript1 = ifElseState("R", AFTER_SCRIPT_2, TEXT);
    Tokenizer.prototype._stateAfterScript2 = ifElseState("I", AFTER_SCRIPT_3, TEXT);
    Tokenizer.prototype._stateAfterScript3 = ifElseState("P", AFTER_SCRIPT_4, TEXT);
    Tokenizer.prototype._stateAfterScript4 = ifElseState("T", AFTER_SCRIPT_5, TEXT);
    Tokenizer.prototype._stateAfterScript5 = function(c) {
      if (c === ">" || whitespace(c)) {
        this._special = SPECIAL_NONE;
        this._state = IN_CLOSING_TAG_NAME;
        this._sectionStart = this._index - 6;
        this._index--;
      } else this._state = TEXT;
    };
    Tokenizer.prototype._stateBeforeStyle1 = consumeSpecialNameChar(
      "Y",
      BEFORE_STYLE_2
    );
    Tokenizer.prototype._stateBeforeStyle2 = consumeSpecialNameChar(
      "L",
      BEFORE_STYLE_3
    );
    Tokenizer.prototype._stateBeforeStyle3 = consumeSpecialNameChar(
      "E",
      BEFORE_STYLE_4
    );
    Tokenizer.prototype._stateBeforeStyle4 = function(c) {
      if (c === "/" || c === ">" || whitespace(c)) {
        this._special = SPECIAL_STYLE;
      }
      this._state = IN_TAG_NAME;
      this._index--;
    };
    Tokenizer.prototype._stateAfterStyle1 = ifElseState("Y", AFTER_STYLE_2, TEXT);
    Tokenizer.prototype._stateAfterStyle2 = ifElseState("L", AFTER_STYLE_3, TEXT);
    Tokenizer.prototype._stateAfterStyle3 = ifElseState("E", AFTER_STYLE_4, TEXT);
    Tokenizer.prototype._stateAfterStyle4 = function(c) {
      if (c === ">" || whitespace(c)) {
        this._special = SPECIAL_NONE;
        this._state = IN_CLOSING_TAG_NAME;
        this._sectionStart = this._index - 5;
        this._index--;
      } else this._state = TEXT;
    };
    Tokenizer.prototype._stateBeforeEntity = ifElseState(
      "#",
      BEFORE_NUMERIC_ENTITY,
      IN_NAMED_ENTITY
    );
    Tokenizer.prototype._stateBeforeNumericEntity = ifElseState(
      "X",
      IN_HEX_ENTITY,
      IN_NUMERIC_ENTITY
    );
    Tokenizer.prototype._parseNamedEntityStrict = function() {
      if (this._sectionStart + 1 < this._index) {
        var entity = this._buffer.substring(
          this._sectionStart + 1,
          this._index
        ), map = this._xmlMode ? xmlMap : entityMap;
        if (map.hasOwnProperty(entity)) {
          this._emitPartial(map[entity]);
          this._sectionStart = this._index + 1;
        }
      }
    };
    Tokenizer.prototype._parseLegacyEntity = function() {
      var start = this._sectionStart + 1, limit = this._index - start;
      if (limit > 6) limit = 6;
      while (limit >= 2) {
        var entity = this._buffer.substr(start, limit);
        if (legacyMap.hasOwnProperty(entity)) {
          this._emitPartial(legacyMap[entity]);
          this._sectionStart += limit + 1;
          return;
        } else {
          limit--;
        }
      }
    };
    Tokenizer.prototype._stateInNamedEntity = function(c) {
      if (c === ";") {
        this._parseNamedEntityStrict();
        if (this._sectionStart + 1 < this._index && !this._xmlMode) {
          this._parseLegacyEntity();
        }
        this._state = this._baseState;
      } else if ((c < "a" || c > "z") && (c < "A" || c > "Z") && (c < "0" || c > "9")) {
        if (this._xmlMode) ;
        else if (this._sectionStart + 1 === this._index) ;
        else if (this._baseState !== TEXT) {
          if (c !== "=") {
            this._parseNamedEntityStrict();
          }
        } else {
          this._parseLegacyEntity();
        }
        this._state = this._baseState;
        this._index--;
      }
    };
    Tokenizer.prototype._decodeNumericEntity = function(offset, base) {
      var sectionStart = this._sectionStart + offset;
      if (sectionStart !== this._index) {
        var entity = this._buffer.substring(sectionStart, this._index);
        var parsed = parseInt(entity, base);
        this._emitPartial(decodeCodePoint(parsed));
        this._sectionStart = this._index;
      } else {
        this._sectionStart--;
      }
      this._state = this._baseState;
    };
    Tokenizer.prototype._stateInNumericEntity = function(c) {
      if (c === ";") {
        this._decodeNumericEntity(2, 10);
        this._sectionStart++;
      } else if (c < "0" || c > "9") {
        if (!this._xmlMode) {
          this._decodeNumericEntity(2, 10);
        } else {
          this._state = this._baseState;
        }
        this._index--;
      }
    };
    Tokenizer.prototype._stateInHexEntity = function(c) {
      if (c === ";") {
        this._decodeNumericEntity(3, 16);
        this._sectionStart++;
      } else if ((c < "a" || c > "f") && (c < "A" || c > "F") && (c < "0" || c > "9")) {
        if (!this._xmlMode) {
          this._decodeNumericEntity(3, 16);
        } else {
          this._state = this._baseState;
        }
        this._index--;
      }
    };
    Tokenizer.prototype._cleanup = function() {
      if (this._sectionStart < 0) {
        this._buffer = "";
        this._bufferOffset += this._index;
        this._index = 0;
      } else if (this._running) {
        if (this._state === TEXT) {
          if (this._sectionStart !== this._index) {
            this._cbs.ontext(this._buffer.substr(this._sectionStart));
          }
          this._buffer = "";
          this._bufferOffset += this._index;
          this._index = 0;
        } else if (this._sectionStart === this._index) {
          this._buffer = "";
          this._bufferOffset += this._index;
          this._index = 0;
        } else {
          this._buffer = this._buffer.substr(this._sectionStart);
          this._index -= this._sectionStart;
          this._bufferOffset += this._sectionStart;
        }
        this._sectionStart = 0;
      }
    };
    Tokenizer.prototype.write = function(chunk) {
      if (this._ended) this._cbs.onerror(Error(".write() after done!"));
      this._buffer += chunk;
      this._parse();
    };
    Tokenizer.prototype._parse = function() {
      while (this._index < this._buffer.length && this._running) {
        var c = this._buffer.charAt(this._index);
        if (this._state === TEXT) {
          this._stateText(c);
        } else if (this._state === BEFORE_TAG_NAME) {
          this._stateBeforeTagName(c);
        } else if (this._state === IN_TAG_NAME) {
          this._stateInTagName(c);
        } else if (this._state === BEFORE_CLOSING_TAG_NAME) {
          this._stateBeforeCloseingTagName(c);
        } else if (this._state === IN_CLOSING_TAG_NAME) {
          this._stateInCloseingTagName(c);
        } else if (this._state === AFTER_CLOSING_TAG_NAME) {
          this._stateAfterCloseingTagName(c);
        } else if (this._state === IN_SELF_CLOSING_TAG) {
          this._stateInSelfClosingTag(c);
        } else if (this._state === BEFORE_ATTRIBUTE_NAME) {
          this._stateBeforeAttributeName(c);
        } else if (this._state === IN_ATTRIBUTE_NAME) {
          this._stateInAttributeName(c);
        } else if (this._state === AFTER_ATTRIBUTE_NAME) {
          this._stateAfterAttributeName(c);
        } else if (this._state === BEFORE_ATTRIBUTE_VALUE) {
          this._stateBeforeAttributeValue(c);
        } else if (this._state === IN_ATTRIBUTE_VALUE_DQ) {
          this._stateInAttributeValueDoubleQuotes(c);
        } else if (this._state === IN_ATTRIBUTE_VALUE_SQ) {
          this._stateInAttributeValueSingleQuotes(c);
        } else if (this._state === IN_ATTRIBUTE_VALUE_NQ) {
          this._stateInAttributeValueNoQuotes(c);
        } else if (this._state === BEFORE_DECLARATION) {
          this._stateBeforeDeclaration(c);
        } else if (this._state === IN_DECLARATION) {
          this._stateInDeclaration(c);
        } else if (this._state === IN_PROCESSING_INSTRUCTION) {
          this._stateInProcessingInstruction(c);
        } else if (this._state === BEFORE_COMMENT) {
          this._stateBeforeComment(c);
        } else if (this._state === IN_COMMENT) {
          this._stateInComment(c);
        } else if (this._state === AFTER_COMMENT_1) {
          this._stateAfterComment1(c);
        } else if (this._state === AFTER_COMMENT_2) {
          this._stateAfterComment2(c);
        } else if (this._state === BEFORE_CDATA_1) {
          this._stateBeforeCdata1(c);
        } else if (this._state === BEFORE_CDATA_2) {
          this._stateBeforeCdata2(c);
        } else if (this._state === BEFORE_CDATA_3) {
          this._stateBeforeCdata3(c);
        } else if (this._state === BEFORE_CDATA_4) {
          this._stateBeforeCdata4(c);
        } else if (this._state === BEFORE_CDATA_5) {
          this._stateBeforeCdata5(c);
        } else if (this._state === BEFORE_CDATA_6) {
          this._stateBeforeCdata6(c);
        } else if (this._state === IN_CDATA) {
          this._stateInCdata(c);
        } else if (this._state === AFTER_CDATA_1) {
          this._stateAfterCdata1(c);
        } else if (this._state === AFTER_CDATA_2) {
          this._stateAfterCdata2(c);
        } else if (this._state === BEFORE_SPECIAL) {
          this._stateBeforeSpecial(c);
        } else if (this._state === BEFORE_SPECIAL_END) {
          this._stateBeforeSpecialEnd(c);
        } else if (this._state === BEFORE_SCRIPT_1) {
          this._stateBeforeScript1(c);
        } else if (this._state === BEFORE_SCRIPT_2) {
          this._stateBeforeScript2(c);
        } else if (this._state === BEFORE_SCRIPT_3) {
          this._stateBeforeScript3(c);
        } else if (this._state === BEFORE_SCRIPT_4) {
          this._stateBeforeScript4(c);
        } else if (this._state === BEFORE_SCRIPT_5) {
          this._stateBeforeScript5(c);
        } else if (this._state === AFTER_SCRIPT_1) {
          this._stateAfterScript1(c);
        } else if (this._state === AFTER_SCRIPT_2) {
          this._stateAfterScript2(c);
        } else if (this._state === AFTER_SCRIPT_3) {
          this._stateAfterScript3(c);
        } else if (this._state === AFTER_SCRIPT_4) {
          this._stateAfterScript4(c);
        } else if (this._state === AFTER_SCRIPT_5) {
          this._stateAfterScript5(c);
        } else if (this._state === BEFORE_STYLE_1) {
          this._stateBeforeStyle1(c);
        } else if (this._state === BEFORE_STYLE_2) {
          this._stateBeforeStyle2(c);
        } else if (this._state === BEFORE_STYLE_3) {
          this._stateBeforeStyle3(c);
        } else if (this._state === BEFORE_STYLE_4) {
          this._stateBeforeStyle4(c);
        } else if (this._state === AFTER_STYLE_1) {
          this._stateAfterStyle1(c);
        } else if (this._state === AFTER_STYLE_2) {
          this._stateAfterStyle2(c);
        } else if (this._state === AFTER_STYLE_3) {
          this._stateAfterStyle3(c);
        } else if (this._state === AFTER_STYLE_4) {
          this._stateAfterStyle4(c);
        } else if (this._state === BEFORE_ENTITY) {
          this._stateBeforeEntity(c);
        } else if (this._state === BEFORE_NUMERIC_ENTITY) {
          this._stateBeforeNumericEntity(c);
        } else if (this._state === IN_NAMED_ENTITY) {
          this._stateInNamedEntity(c);
        } else if (this._state === IN_NUMERIC_ENTITY) {
          this._stateInNumericEntity(c);
        } else if (this._state === IN_HEX_ENTITY) {
          this._stateInHexEntity(c);
        } else {
          this._cbs.onerror(Error("unknown _state"), this._state);
        }
        this._index++;
      }
      this._cleanup();
    };
    Tokenizer.prototype.pause = function() {
      this._running = false;
    };
    Tokenizer.prototype.resume = function() {
      this._running = true;
      if (this._index < this._buffer.length) {
        this._parse();
      }
      if (this._ended) {
        this._finish();
      }
    };
    Tokenizer.prototype.end = function(chunk) {
      if (this._ended) this._cbs.onerror(Error(".end() after done!"));
      if (chunk) this.write(chunk);
      this._ended = true;
      if (this._running) this._finish();
    };
    Tokenizer.prototype._finish = function() {
      if (this._sectionStart < this._index) {
        this._handleTrailingData();
      }
      this._cbs.onend();
    };
    Tokenizer.prototype._handleTrailingData = function() {
      var data = this._buffer.substr(this._sectionStart);
      if (this._state === IN_CDATA || this._state === AFTER_CDATA_1 || this._state === AFTER_CDATA_2) {
        this._cbs.oncdata(data);
      } else if (this._state === IN_COMMENT || this._state === AFTER_COMMENT_1 || this._state === AFTER_COMMENT_2) {
        this._cbs.oncomment(data);
      } else if (this._state === IN_NAMED_ENTITY && !this._xmlMode) {
        this._parseLegacyEntity();
        if (this._sectionStart < this._index) {
          this._state = this._baseState;
          this._handleTrailingData();
        }
      } else if (this._state === IN_NUMERIC_ENTITY && !this._xmlMode) {
        this._decodeNumericEntity(2, 10);
        if (this._sectionStart < this._index) {
          this._state = this._baseState;
          this._handleTrailingData();
        }
      } else if (this._state === IN_HEX_ENTITY && !this._xmlMode) {
        this._decodeNumericEntity(3, 16);
        if (this._sectionStart < this._index) {
          this._state = this._baseState;
          this._handleTrailingData();
        }
      } else if (this._state !== IN_TAG_NAME && this._state !== BEFORE_ATTRIBUTE_NAME && this._state !== BEFORE_ATTRIBUTE_VALUE && this._state !== AFTER_ATTRIBUTE_NAME && this._state !== IN_ATTRIBUTE_NAME && this._state !== IN_ATTRIBUTE_VALUE_SQ && this._state !== IN_ATTRIBUTE_VALUE_DQ && this._state !== IN_ATTRIBUTE_VALUE_NQ && this._state !== IN_CLOSING_TAG_NAME) {
        this._cbs.ontext(data);
      }
    };
    Tokenizer.prototype.reset = function() {
      Tokenizer.call(
        this,
        { xmlMode: this._xmlMode, decodeEntities: this._decodeEntities },
        this._cbs
      );
    };
    Tokenizer.prototype.getAbsoluteIndex = function() {
      return this._bufferOffset + this._index;
    };
    Tokenizer.prototype._getSection = function() {
      return this._buffer.substring(this._sectionStart, this._index);
    };
    Tokenizer.prototype._emitToken = function(name) {
      this._cbs[name](this._getSection());
      this._sectionStart = -1;
    };
    Tokenizer.prototype._emitPartial = function(value) {
      if (this._baseState !== TEXT) {
        this._cbs.onattribdata(value);
      } else {
        this._cbs.ontext(value);
      }
    };
  }
});

// node_modules/inherits/inherits_browser.js
var require_inherits_browser = __commonJS({
  "node_modules/inherits/inherits_browser.js"(exports2, module2) {
    init_process_shim();
    if (typeof Object.create === "function") {
      module2.exports = function inherits(ctor, superCtor) {
        if (superCtor) {
          ctor.super_ = superCtor;
          ctor.prototype = Object.create(superCtor.prototype, {
            constructor: {
              value: ctor,
              enumerable: false,
              writable: true,
              configurable: true
            }
          });
        }
      };
    } else {
      module2.exports = function inherits(ctor, superCtor) {
        if (superCtor) {
          ctor.super_ = superCtor;
          var TempCtor = function() {
          };
          TempCtor.prototype = superCtor.prototype;
          ctor.prototype = new TempCtor();
          ctor.prototype.constructor = ctor;
        }
      };
    }
  }
});

// node_modules/htmlparser2/lib/Parser.js
var require_Parser = __commonJS({
  "node_modules/htmlparser2/lib/Parser.js"(exports2, module2) {
    init_process_shim();
    var Tokenizer = require_Tokenizer();
    var formTags = {
      input: true,
      option: true,
      optgroup: true,
      select: true,
      button: true,
      datalist: true,
      textarea: true
    };
    var openImpliesClose = {
      tr: { tr: true, th: true, td: true },
      th: { th: true },
      td: { thead: true, th: true, td: true },
      body: { head: true, link: true, script: true },
      li: { li: true },
      p: { p: true },
      h1: { p: true },
      h2: { p: true },
      h3: { p: true },
      h4: { p: true },
      h5: { p: true },
      h6: { p: true },
      select: formTags,
      input: formTags,
      output: formTags,
      button: formTags,
      datalist: formTags,
      textarea: formTags,
      option: { option: true },
      optgroup: { optgroup: true }
    };
    var voidElements = {
      __proto__: null,
      area: true,
      base: true,
      basefont: true,
      br: true,
      col: true,
      command: true,
      embed: true,
      frame: true,
      hr: true,
      img: true,
      input: true,
      isindex: true,
      keygen: true,
      link: true,
      meta: true,
      param: true,
      source: true,
      track: true,
      wbr: true
    };
    var foreignContextElements = {
      __proto__: null,
      math: true,
      svg: true
    };
    var htmlIntegrationElements = {
      __proto__: null,
      mi: true,
      mo: true,
      mn: true,
      ms: true,
      mtext: true,
      "annotation-xml": true,
      foreignObject: true,
      desc: true,
      title: true
    };
    var re_nameEnd = /\s|\//;
    function Parser(cbs, options) {
      this._options = options || {};
      this._cbs = cbs || {};
      this._tagname = "";
      this._attribname = "";
      this._attribvalue = "";
      this._attribs = null;
      this._stack = [];
      this._foreignContext = [];
      this.startIndex = 0;
      this.endIndex = null;
      this._lowerCaseTagNames = "lowerCaseTags" in this._options ? !!this._options.lowerCaseTags : !this._options.xmlMode;
      this._lowerCaseAttributeNames = "lowerCaseAttributeNames" in this._options ? !!this._options.lowerCaseAttributeNames : !this._options.xmlMode;
      if (this._options.Tokenizer) {
        Tokenizer = this._options.Tokenizer;
      }
      this._tokenizer = new Tokenizer(this._options, this);
      if (this._cbs.onparserinit) this._cbs.onparserinit(this);
    }
    require_inherits_browser()(Parser, require("events").EventEmitter);
    Parser.prototype._updatePosition = function(initialOffset) {
      if (this.endIndex === null) {
        if (this._tokenizer._sectionStart <= initialOffset) {
          this.startIndex = 0;
        } else {
          this.startIndex = this._tokenizer._sectionStart - initialOffset;
        }
      } else this.startIndex = this.endIndex + 1;
      this.endIndex = this._tokenizer.getAbsoluteIndex();
    };
    Parser.prototype.ontext = function(data) {
      this._updatePosition(1);
      this.endIndex--;
      if (this._cbs.ontext) this._cbs.ontext(data);
    };
    Parser.prototype.onopentagname = function(name) {
      if (this._lowerCaseTagNames) {
        name = name.toLowerCase();
      }
      this._tagname = name;
      if (!this._options.xmlMode && name in openImpliesClose) {
        for (var el; (el = this._stack[this._stack.length - 1]) in openImpliesClose[name]; this.onclosetag(el)) ;
      }
      if (this._options.xmlMode || !(name in voidElements)) {
        this._stack.push(name);
        if (name in foreignContextElements) this._foreignContext.push(true);
        else if (name in htmlIntegrationElements)
          this._foreignContext.push(false);
      }
      if (this._cbs.onopentagname) this._cbs.onopentagname(name);
      if (this._cbs.onopentag) this._attribs = {};
    };
    Parser.prototype.onopentagend = function() {
      this._updatePosition(1);
      if (this._attribs) {
        if (this._cbs.onopentag)
          this._cbs.onopentag(this._tagname, this._attribs);
        this._attribs = null;
      }
      if (!this._options.xmlMode && this._cbs.onclosetag && this._tagname in voidElements) {
        this._cbs.onclosetag(this._tagname);
      }
      this._tagname = "";
    };
    Parser.prototype.onclosetag = function(name) {
      this._updatePosition(1);
      if (this._lowerCaseTagNames) {
        name = name.toLowerCase();
      }
      if (name in foreignContextElements || name in htmlIntegrationElements) {
        this._foreignContext.pop();
      }
      if (this._stack.length && (!(name in voidElements) || this._options.xmlMode)) {
        var pos = this._stack.lastIndexOf(name);
        if (pos !== -1) {
          if (this._cbs.onclosetag) {
            pos = this._stack.length - pos;
            while (pos--) this._cbs.onclosetag(this._stack.pop());
          } else this._stack.length = pos;
        } else if (name === "p" && !this._options.xmlMode) {
          this.onopentagname(name);
          this._closeCurrentTag();
        }
      } else if (!this._options.xmlMode && (name === "br" || name === "p")) {
        this.onopentagname(name);
        this._closeCurrentTag();
      }
    };
    Parser.prototype.onselfclosingtag = function() {
      if (this._options.xmlMode || this._options.recognizeSelfClosing || this._foreignContext[this._foreignContext.length - 1]) {
        this._closeCurrentTag();
      } else {
        this.onopentagend();
      }
    };
    Parser.prototype._closeCurrentTag = function() {
      var name = this._tagname;
      this.onopentagend();
      if (this._stack[this._stack.length - 1] === name) {
        if (this._cbs.onclosetag) {
          this._cbs.onclosetag(name);
        }
        this._stack.pop();
      }
    };
    Parser.prototype.onattribname = function(name) {
      if (this._lowerCaseAttributeNames) {
        name = name.toLowerCase();
      }
      this._attribname = name;
    };
    Parser.prototype.onattribdata = function(value) {
      this._attribvalue += value;
    };
    Parser.prototype.onattribend = function() {
      if (this._cbs.onattribute)
        this._cbs.onattribute(this._attribname, this._attribvalue);
      if (this._attribs && !Object.prototype.hasOwnProperty.call(this._attribs, this._attribname)) {
        this._attribs[this._attribname] = this._attribvalue;
      }
      this._attribname = "";
      this._attribvalue = "";
    };
    Parser.prototype._getInstructionName = function(value) {
      var idx = value.search(re_nameEnd), name = idx < 0 ? value : value.substr(0, idx);
      if (this._lowerCaseTagNames) {
        name = name.toLowerCase();
      }
      return name;
    };
    Parser.prototype.ondeclaration = function(value) {
      if (this._cbs.onprocessinginstruction) {
        var name = this._getInstructionName(value);
        this._cbs.onprocessinginstruction("!" + name, "!" + value);
      }
    };
    Parser.prototype.onprocessinginstruction = function(value) {
      if (this._cbs.onprocessinginstruction) {
        var name = this._getInstructionName(value);
        this._cbs.onprocessinginstruction("?" + name, "?" + value);
      }
    };
    Parser.prototype.oncomment = function(value) {
      this._updatePosition(4);
      if (this._cbs.oncomment) this._cbs.oncomment(value);
      if (this._cbs.oncommentend) this._cbs.oncommentend();
    };
    Parser.prototype.oncdata = function(value) {
      this._updatePosition(1);
      if (this._options.xmlMode || this._options.recognizeCDATA) {
        if (this._cbs.oncdatastart) this._cbs.oncdatastart();
        if (this._cbs.ontext) this._cbs.ontext(value);
        if (this._cbs.oncdataend) this._cbs.oncdataend();
      } else {
        this.oncomment("[CDATA[" + value + "]]");
      }
    };
    Parser.prototype.onerror = function(err) {
      if (this._cbs.onerror) this._cbs.onerror(err);
    };
    Parser.prototype.onend = function() {
      if (this._cbs.onclosetag) {
        for (var i = this._stack.length; i > 0; this._cbs.onclosetag(this._stack[--i])) ;
      }
      if (this._cbs.onend) this._cbs.onend();
    };
    Parser.prototype.reset = function() {
      if (this._cbs.onreset) this._cbs.onreset();
      this._tokenizer.reset();
      this._tagname = "";
      this._attribname = "";
      this._attribs = null;
      this._stack = [];
      if (this._cbs.onparserinit) this._cbs.onparserinit(this);
    };
    Parser.prototype.parseComplete = function(data) {
      this.reset();
      this.end(data);
    };
    Parser.prototype.write = function(chunk) {
      this._tokenizer.write(chunk);
    };
    Parser.prototype.end = function(chunk) {
      this._tokenizer.end(chunk);
    };
    Parser.prototype.pause = function() {
      this._tokenizer.pause();
    };
    Parser.prototype.resume = function() {
      this._tokenizer.resume();
    };
    Parser.prototype.parseChunk = Parser.prototype.write;
    Parser.prototype.done = Parser.prototype.end;
    module2.exports = Parser;
  }
});

// node_modules/domelementtype/index.js
var require_domelementtype = __commonJS({
  "node_modules/domelementtype/index.js"(exports2, module2) {
    init_process_shim();
    module2.exports = {
      Text: "text",
      //Text
      Directive: "directive",
      //<? ... ?>
      Comment: "comment",
      //<!-- ... -->
      Script: "script",
      //<script> tags
      Style: "style",
      //<style> tags
      Tag: "tag",
      //Any tag
      CDATA: "cdata",
      //<![CDATA[ ... ]]>
      Doctype: "doctype",
      isTag: function(elem) {
        return elem.type === "tag" || elem.type === "script" || elem.type === "style";
      }
    };
  }
});

// node_modules/domhandler/lib/node.js
var require_node = __commonJS({
  "node_modules/domhandler/lib/node.js"(exports2, module2) {
    init_process_shim();
    var NodePrototype = module2.exports = {
      get firstChild() {
        var children = this.children;
        return children && children[0] || null;
      },
      get lastChild() {
        var children = this.children;
        return children && children[children.length - 1] || null;
      },
      get nodeType() {
        return nodeTypes[this.type] || nodeTypes.element;
      }
    };
    var domLvl1 = {
      tagName: "name",
      childNodes: "children",
      parentNode: "parent",
      previousSibling: "prev",
      nextSibling: "next",
      nodeValue: "data"
    };
    var nodeTypes = {
      element: 1,
      text: 3,
      cdata: 4,
      comment: 8
    };
    Object.keys(domLvl1).forEach(function(key) {
      var shorthand = domLvl1[key];
      Object.defineProperty(NodePrototype, key, {
        get: function() {
          return this[shorthand] || null;
        },
        set: function(val) {
          this[shorthand] = val;
          return val;
        }
      });
    });
  }
});

// node_modules/domhandler/lib/element.js
var require_element = __commonJS({
  "node_modules/domhandler/lib/element.js"(exports2, module2) {
    init_process_shim();
    var NodePrototype = require_node();
    var ElementPrototype = module2.exports = Object.create(NodePrototype);
    var domLvl1 = {
      tagName: "name"
    };
    Object.keys(domLvl1).forEach(function(key) {
      var shorthand = domLvl1[key];
      Object.defineProperty(ElementPrototype, key, {
        get: function() {
          return this[shorthand] || null;
        },
        set: function(val) {
          this[shorthand] = val;
          return val;
        }
      });
    });
  }
});

// node_modules/domhandler/index.js
var require_domhandler = __commonJS({
  "node_modules/domhandler/index.js"(exports2, module2) {
    init_process_shim();
    var ElementType = require_domelementtype();
    var re_whitespace = /\s+/g;
    var NodePrototype = require_node();
    var ElementPrototype = require_element();
    function DomHandler(callback, options, elementCB) {
      if (typeof callback === "object") {
        elementCB = options;
        options = callback;
        callback = null;
      } else if (typeof options === "function") {
        elementCB = options;
        options = defaultOpts;
      }
      this._callback = callback;
      this._options = options || defaultOpts;
      this._elementCB = elementCB;
      this.dom = [];
      this._done = false;
      this._tagStack = [];
      this._parser = this._parser || null;
    }
    var defaultOpts = {
      normalizeWhitespace: false,
      //Replace all whitespace with single spaces
      withStartIndices: false,
      //Add startIndex properties to nodes
      withEndIndices: false
      //Add endIndex properties to nodes
    };
    DomHandler.prototype.onparserinit = function(parser) {
      this._parser = parser;
    };
    DomHandler.prototype.onreset = function() {
      DomHandler.call(this, this._callback, this._options, this._elementCB);
    };
    DomHandler.prototype.onend = function() {
      if (this._done) return;
      this._done = true;
      this._parser = null;
      this._handleCallback(null);
    };
    DomHandler.prototype._handleCallback = DomHandler.prototype.onerror = function(error) {
      if (typeof this._callback === "function") {
        this._callback(error, this.dom);
      } else {
        if (error) throw error;
      }
    };
    DomHandler.prototype.onclosetag = function() {
      var elem = this._tagStack.pop();
      if (this._options.withEndIndices && elem) {
        elem.endIndex = this._parser.endIndex;
      }
      if (this._elementCB) this._elementCB(elem);
    };
    DomHandler.prototype._createDomElement = function(properties) {
      if (!this._options.withDomLvl1) return properties;
      var element;
      if (properties.type === "tag") {
        element = Object.create(ElementPrototype);
      } else {
        element = Object.create(NodePrototype);
      }
      for (var key in properties) {
        if (properties.hasOwnProperty(key)) {
          element[key] = properties[key];
        }
      }
      return element;
    };
    DomHandler.prototype._addDomElement = function(element) {
      var parent = this._tagStack[this._tagStack.length - 1];
      var siblings = parent ? parent.children : this.dom;
      var previousSibling = siblings[siblings.length - 1];
      element.next = null;
      if (this._options.withStartIndices) {
        element.startIndex = this._parser.startIndex;
      }
      if (this._options.withEndIndices) {
        element.endIndex = this._parser.endIndex;
      }
      if (previousSibling) {
        element.prev = previousSibling;
        previousSibling.next = element;
      } else {
        element.prev = null;
      }
      siblings.push(element);
      element.parent = parent || null;
    };
    DomHandler.prototype.onopentag = function(name, attribs) {
      var properties = {
        type: name === "script" ? ElementType.Script : name === "style" ? ElementType.Style : ElementType.Tag,
        name,
        attribs,
        children: []
      };
      var element = this._createDomElement(properties);
      this._addDomElement(element);
      this._tagStack.push(element);
    };
    DomHandler.prototype.ontext = function(data) {
      var normalize = this._options.normalizeWhitespace || this._options.ignoreWhitespace;
      var lastTag;
      if (!this._tagStack.length && this.dom.length && (lastTag = this.dom[this.dom.length - 1]).type === ElementType.Text) {
        if (normalize) {
          lastTag.data = (lastTag.data + data).replace(re_whitespace, " ");
        } else {
          lastTag.data += data;
        }
      } else {
        if (this._tagStack.length && (lastTag = this._tagStack[this._tagStack.length - 1]) && (lastTag = lastTag.children[lastTag.children.length - 1]) && lastTag.type === ElementType.Text) {
          if (normalize) {
            lastTag.data = (lastTag.data + data).replace(re_whitespace, " ");
          } else {
            lastTag.data += data;
          }
        } else {
          if (normalize) {
            data = data.replace(re_whitespace, " ");
          }
          var element = this._createDomElement({
            data,
            type: ElementType.Text
          });
          this._addDomElement(element);
        }
      }
    };
    DomHandler.prototype.oncomment = function(data) {
      var lastTag = this._tagStack[this._tagStack.length - 1];
      if (lastTag && lastTag.type === ElementType.Comment) {
        lastTag.data += data;
        return;
      }
      var properties = {
        data,
        type: ElementType.Comment
      };
      var element = this._createDomElement(properties);
      this._addDomElement(element);
      this._tagStack.push(element);
    };
    DomHandler.prototype.oncdatastart = function() {
      var properties = {
        children: [{
          data: "",
          type: ElementType.Text
        }],
        type: ElementType.CDATA
      };
      var element = this._createDomElement(properties);
      this._addDomElement(element);
      this._tagStack.push(element);
    };
    DomHandler.prototype.oncommentend = DomHandler.prototype.oncdataend = function() {
      this._tagStack.pop();
    };
    DomHandler.prototype.onprocessinginstruction = function(name, data) {
      var element = this._createDomElement({
        name,
        data,
        type: ElementType.Directive
      });
      this._addDomElement(element);
    };
    module2.exports = DomHandler;
  }
});

// node_modules/entities/lib/encode.js
var require_encode = __commonJS({
  "node_modules/entities/lib/encode.js"(exports2) {
    init_process_shim();
    var inverseXML = getInverseObj(require_xml());
    var xmlReplacer = getInverseReplacer(inverseXML);
    exports2.XML = getInverse(inverseXML, xmlReplacer);
    var inverseHTML = getInverseObj(require_entities());
    var htmlReplacer = getInverseReplacer(inverseHTML);
    exports2.HTML = getInverse(inverseHTML, htmlReplacer);
    function getInverseObj(obj) {
      return Object.keys(obj).sort().reduce(function(inverse, name) {
        inverse[obj[name]] = "&" + name + ";";
        return inverse;
      }, {});
    }
    function getInverseReplacer(inverse) {
      var single = [], multiple = [];
      Object.keys(inverse).forEach(function(k) {
        if (k.length === 1) {
          single.push("\\" + k);
        } else {
          multiple.push(k);
        }
      });
      multiple.unshift("[" + single.join("") + "]");
      return new RegExp(multiple.join("|"), "g");
    }
    var re_nonASCII = /[^\0-\x7F]/g;
    var re_astralSymbols = /[\uD800-\uDBFF][\uDC00-\uDFFF]/g;
    function singleCharReplacer(c) {
      return "&#x" + c.charCodeAt(0).toString(16).toUpperCase() + ";";
    }
    function astralReplacer(c) {
      var high = c.charCodeAt(0);
      var low = c.charCodeAt(1);
      var codePoint = (high - 55296) * 1024 + low - 56320 + 65536;
      return "&#x" + codePoint.toString(16).toUpperCase() + ";";
    }
    function getInverse(inverse, re) {
      function func(name) {
        return inverse[name];
      }
      return function(data) {
        return data.replace(re, func).replace(re_astralSymbols, astralReplacer).replace(re_nonASCII, singleCharReplacer);
      };
    }
    var re_xmlChars = getInverseReplacer(inverseXML);
    function escapeXML(data) {
      return data.replace(re_xmlChars, singleCharReplacer).replace(re_astralSymbols, astralReplacer).replace(re_nonASCII, singleCharReplacer);
    }
    exports2.escape = escapeXML;
  }
});

// node_modules/entities/lib/decode.js
var require_decode2 = __commonJS({
  "node_modules/entities/lib/decode.js"(exports2, module2) {
    init_process_shim();
    var entityMap = require_entities();
    var legacyMap = require_legacy();
    var xmlMap = require_xml();
    var decodeCodePoint = require_decode_codepoint();
    var decodeXMLStrict = getStrictDecoder(xmlMap);
    var decodeHTMLStrict = getStrictDecoder(entityMap);
    function getStrictDecoder(map) {
      var keys = Object.keys(map).join("|"), replace = getReplacer(map);
      keys += "|#[xX][\\da-fA-F]+|#\\d+";
      var re = new RegExp("&(?:" + keys + ");", "g");
      return function(str) {
        return String(str).replace(re, replace);
      };
    }
    var decodeHTML = (function() {
      var legacy = Object.keys(legacyMap).sort(sorter);
      var keys = Object.keys(entityMap).sort(sorter);
      for (var i = 0, j = 0; i < keys.length; i++) {
        if (legacy[j] === keys[i]) {
          keys[i] += ";?";
          j++;
        } else {
          keys[i] += ";";
        }
      }
      var re = new RegExp("&(?:" + keys.join("|") + "|#[xX][\\da-fA-F]+;?|#\\d+;?)", "g"), replace = getReplacer(entityMap);
      function replacer(str) {
        if (str.substr(-1) !== ";") str += ";";
        return replace(str);
      }
      return function(str) {
        return String(str).replace(re, replacer);
      };
    })();
    function sorter(a, b) {
      return a < b ? 1 : -1;
    }
    function getReplacer(map) {
      return function replace(str) {
        if (str.charAt(1) === "#") {
          if (str.charAt(2) === "X" || str.charAt(2) === "x") {
            return decodeCodePoint(parseInt(str.substr(3), 16));
          }
          return decodeCodePoint(parseInt(str.substr(2), 10));
        }
        return map[str.slice(1, -1)];
      };
    }
    module2.exports = {
      XML: decodeXMLStrict,
      HTML: decodeHTML,
      HTMLStrict: decodeHTMLStrict
    };
  }
});

// node_modules/entities/index.js
var require_entities2 = __commonJS({
  "node_modules/entities/index.js"(exports2) {
    init_process_shim();
    var encode = require_encode();
    var decode = require_decode2();
    exports2.decode = function(data, level) {
      return (!level || level <= 0 ? decode.XML : decode.HTML)(data);
    };
    exports2.decodeStrict = function(data, level) {
      return (!level || level <= 0 ? decode.XML : decode.HTMLStrict)(data);
    };
    exports2.encode = function(data, level) {
      return (!level || level <= 0 ? encode.XML : encode.HTML)(data);
    };
    exports2.encodeXML = encode.XML;
    exports2.encodeHTML4 = exports2.encodeHTML5 = exports2.encodeHTML = encode.HTML;
    exports2.decodeXML = exports2.decodeXMLStrict = decode.XML;
    exports2.decodeHTML4 = exports2.decodeHTML5 = exports2.decodeHTML = decode.HTML;
    exports2.decodeHTML4Strict = exports2.decodeHTML5Strict = exports2.decodeHTMLStrict = decode.HTMLStrict;
    exports2.escape = encode.escape;
  }
});

// node_modules/dom-serializer/index.js
var require_dom_serializer = __commonJS({
  "node_modules/dom-serializer/index.js"(exports2, module2) {
    init_process_shim();
    var ElementType = require_domelementtype();
    var entities = require_entities2();
    var unencodedElements = {
      __proto__: null,
      style: true,
      script: true,
      xmp: true,
      iframe: true,
      noembed: true,
      noframes: true,
      plaintext: true,
      noscript: true
    };
    function formatAttrs(attributes, opts) {
      if (!attributes) return;
      var output = "", value;
      for (var key in attributes) {
        value = attributes[key];
        if (output) {
          output += " ";
        }
        output += key;
        if (value !== null && value !== "" || opts.xmlMode) {
          output += '="' + (opts.decodeEntities ? entities.encodeXML(value) : value) + '"';
        }
      }
      return output;
    }
    var singleTag = {
      __proto__: null,
      area: true,
      base: true,
      basefont: true,
      br: true,
      col: true,
      command: true,
      embed: true,
      frame: true,
      hr: true,
      img: true,
      input: true,
      isindex: true,
      keygen: true,
      link: true,
      meta: true,
      param: true,
      source: true,
      track: true,
      wbr: true
    };
    var render = module2.exports = function(dom, opts) {
      if (!Array.isArray(dom) && !dom.cheerio) dom = [dom];
      opts = opts || {};
      var output = "";
      for (var i = 0; i < dom.length; i++) {
        var elem = dom[i];
        if (elem.type === "root")
          output += render(elem.children, opts);
        else if (ElementType.isTag(elem))
          output += renderTag(elem, opts);
        else if (elem.type === ElementType.Directive)
          output += renderDirective(elem);
        else if (elem.type === ElementType.Comment)
          output += renderComment(elem);
        else if (elem.type === ElementType.CDATA)
          output += renderCdata(elem);
        else
          output += renderText(elem, opts);
      }
      return output;
    };
    function renderTag(elem, opts) {
      if (elem.name === "svg") opts = { decodeEntities: opts.decodeEntities, xmlMode: true };
      var tag = "<" + elem.name, attribs = formatAttrs(elem.attribs, opts);
      if (attribs) {
        tag += " " + attribs;
      }
      if (opts.xmlMode && (!elem.children || elem.children.length === 0)) {
        tag += "/>";
      } else {
        tag += ">";
        if (elem.children) {
          tag += render(elem.children, opts);
        }
        if (!singleTag[elem.name] || opts.xmlMode) {
          tag += "</" + elem.name + ">";
        }
      }
      return tag;
    }
    function renderDirective(elem) {
      return "<" + elem.data + ">";
    }
    function renderText(elem, opts) {
      var data = elem.data || "";
      if (opts.decodeEntities && !(elem.parent && elem.parent.name in unencodedElements)) {
        data = entities.encodeXML(data);
      }
      return data;
    }
    function renderCdata(elem) {
      return "<![CDATA[" + elem.children[0].data + "]]>";
    }
    function renderComment(elem) {
      return "<!--" + elem.data + "-->";
    }
  }
});

// node_modules/domutils/lib/stringify.js
var require_stringify = __commonJS({
  "node_modules/domutils/lib/stringify.js"(exports2, module2) {
    init_process_shim();
    var ElementType = require_domelementtype();
    var getOuterHTML = require_dom_serializer();
    var isTag = ElementType.isTag;
    module2.exports = {
      getInnerHTML,
      getOuterHTML,
      getText
    };
    function getInnerHTML(elem, opts) {
      return elem.children ? elem.children.map(function(elem2) {
        return getOuterHTML(elem2, opts);
      }).join("") : "";
    }
    function getText(elem) {
      if (Array.isArray(elem)) return elem.map(getText).join("");
      if (isTag(elem) || elem.type === ElementType.CDATA) return getText(elem.children);
      if (elem.type === ElementType.Text) return elem.data;
      return "";
    }
  }
});

// node_modules/domutils/lib/traversal.js
var require_traversal = __commonJS({
  "node_modules/domutils/lib/traversal.js"(exports2) {
    init_process_shim();
    var getChildren = exports2.getChildren = function(elem) {
      return elem.children;
    };
    var getParent = exports2.getParent = function(elem) {
      return elem.parent;
    };
    exports2.getSiblings = function(elem) {
      var parent = getParent(elem);
      return parent ? getChildren(parent) : [elem];
    };
    exports2.getAttributeValue = function(elem, name) {
      return elem.attribs && elem.attribs[name];
    };
    exports2.hasAttrib = function(elem, name) {
      return !!elem.attribs && hasOwnProperty.call(elem.attribs, name);
    };
    exports2.getName = function(elem) {
      return elem.name;
    };
  }
});

// node_modules/domutils/lib/manipulation.js
var require_manipulation = __commonJS({
  "node_modules/domutils/lib/manipulation.js"(exports2) {
    init_process_shim();
    exports2.removeElement = function(elem) {
      if (elem.prev) elem.prev.next = elem.next;
      if (elem.next) elem.next.prev = elem.prev;
      if (elem.parent) {
        var childs = elem.parent.children;
        childs.splice(childs.lastIndexOf(elem), 1);
      }
    };
    exports2.replaceElement = function(elem, replacement) {
      var prev = replacement.prev = elem.prev;
      if (prev) {
        prev.next = replacement;
      }
      var next = replacement.next = elem.next;
      if (next) {
        next.prev = replacement;
      }
      var parent = replacement.parent = elem.parent;
      if (parent) {
        var childs = parent.children;
        childs[childs.lastIndexOf(elem)] = replacement;
      }
    };
    exports2.appendChild = function(elem, child) {
      child.parent = elem;
      if (elem.children.push(child) !== 1) {
        var sibling = elem.children[elem.children.length - 2];
        sibling.next = child;
        child.prev = sibling;
        child.next = null;
      }
    };
    exports2.append = function(elem, next) {
      var parent = elem.parent, currNext = elem.next;
      next.next = currNext;
      next.prev = elem;
      elem.next = next;
      next.parent = parent;
      if (currNext) {
        currNext.prev = next;
        if (parent) {
          var childs = parent.children;
          childs.splice(childs.lastIndexOf(currNext), 0, next);
        }
      } else if (parent) {
        parent.children.push(next);
      }
    };
    exports2.prepend = function(elem, prev) {
      var parent = elem.parent;
      if (parent) {
        var childs = parent.children;
        childs.splice(childs.lastIndexOf(elem), 0, prev);
      }
      if (elem.prev) {
        elem.prev.next = prev;
      }
      prev.parent = parent;
      prev.prev = elem.prev;
      prev.next = elem;
      elem.prev = prev;
    };
  }
});

// node_modules/domutils/lib/querying.js
var require_querying = __commonJS({
  "node_modules/domutils/lib/querying.js"(exports2, module2) {
    init_process_shim();
    var isTag = require_domelementtype().isTag;
    module2.exports = {
      filter,
      find,
      findOneChild,
      findOne,
      existsOne,
      findAll
    };
    function filter(test, element, recurse, limit) {
      if (!Array.isArray(element)) element = [element];
      if (typeof limit !== "number" || !isFinite(limit)) {
        limit = Infinity;
      }
      return find(test, element, recurse !== false, limit);
    }
    function find(test, elems, recurse, limit) {
      var result = [], childs;
      for (var i = 0, j = elems.length; i < j; i++) {
        if (test(elems[i])) {
          result.push(elems[i]);
          if (--limit <= 0) break;
        }
        childs = elems[i].children;
        if (recurse && childs && childs.length > 0) {
          childs = find(test, childs, recurse, limit);
          result = result.concat(childs);
          limit -= childs.length;
          if (limit <= 0) break;
        }
      }
      return result;
    }
    function findOneChild(test, elems) {
      for (var i = 0, l = elems.length; i < l; i++) {
        if (test(elems[i])) return elems[i];
      }
      return null;
    }
    function findOne(test, elems) {
      var elem = null;
      for (var i = 0, l = elems.length; i < l && !elem; i++) {
        if (!isTag(elems[i])) {
          continue;
        } else if (test(elems[i])) {
          elem = elems[i];
        } else if (elems[i].children.length > 0) {
          elem = findOne(test, elems[i].children);
        }
      }
      return elem;
    }
    function existsOne(test, elems) {
      for (var i = 0, l = elems.length; i < l; i++) {
        if (isTag(elems[i]) && (test(elems[i]) || elems[i].children.length > 0 && existsOne(test, elems[i].children))) {
          return true;
        }
      }
      return false;
    }
    function findAll(test, elems) {
      var result = [];
      for (var i = 0, j = elems.length; i < j; i++) {
        if (!isTag(elems[i])) continue;
        if (test(elems[i])) result.push(elems[i]);
        if (elems[i].children.length > 0) {
          result = result.concat(findAll(test, elems[i].children));
        }
      }
      return result;
    }
  }
});

// node_modules/domutils/lib/legacy.js
var require_legacy2 = __commonJS({
  "node_modules/domutils/lib/legacy.js"(exports2) {
    init_process_shim();
    var ElementType = require_domelementtype();
    var isTag = exports2.isTag = ElementType.isTag;
    exports2.testElement = function(options, element) {
      for (var key in options) {
        if (!options.hasOwnProperty(key)) ;
        else if (key === "tag_name") {
          if (!isTag(element) || !options.tag_name(element.name)) {
            return false;
          }
        } else if (key === "tag_type") {
          if (!options.tag_type(element.type)) return false;
        } else if (key === "tag_contains") {
          if (isTag(element) || !options.tag_contains(element.data)) {
            return false;
          }
        } else if (!element.attribs || !options[key](element.attribs[key])) {
          return false;
        }
      }
      return true;
    };
    var Checks = {
      tag_name: function(name) {
        if (typeof name === "function") {
          return function(elem) {
            return isTag(elem) && name(elem.name);
          };
        } else if (name === "*") {
          return isTag;
        } else {
          return function(elem) {
            return isTag(elem) && elem.name === name;
          };
        }
      },
      tag_type: function(type) {
        if (typeof type === "function") {
          return function(elem) {
            return type(elem.type);
          };
        } else {
          return function(elem) {
            return elem.type === type;
          };
        }
      },
      tag_contains: function(data) {
        if (typeof data === "function") {
          return function(elem) {
            return !isTag(elem) && data(elem.data);
          };
        } else {
          return function(elem) {
            return !isTag(elem) && elem.data === data;
          };
        }
      }
    };
    function getAttribCheck(attrib, value) {
      if (typeof value === "function") {
        return function(elem) {
          return elem.attribs && value(elem.attribs[attrib]);
        };
      } else {
        return function(elem) {
          return elem.attribs && elem.attribs[attrib] === value;
        };
      }
    }
    function combineFuncs(a, b) {
      return function(elem) {
        return a(elem) || b(elem);
      };
    }
    exports2.getElements = function(options, element, recurse, limit) {
      var funcs = Object.keys(options).map(function(key) {
        var value = options[key];
        return key in Checks ? Checks[key](value) : getAttribCheck(key, value);
      });
      return funcs.length === 0 ? [] : this.filter(
        funcs.reduce(combineFuncs),
        element,
        recurse,
        limit
      );
    };
    exports2.getElementById = function(id, element, recurse) {
      if (!Array.isArray(element)) element = [element];
      return this.findOne(getAttribCheck("id", id), element, recurse !== false);
    };
    exports2.getElementsByTagName = function(name, element, recurse, limit) {
      return this.filter(Checks.tag_name(name), element, recurse, limit);
    };
    exports2.getElementsByTagType = function(type, element, recurse, limit) {
      return this.filter(Checks.tag_type(type), element, recurse, limit);
    };
  }
});

// node_modules/domutils/lib/helpers.js
var require_helpers = __commonJS({
  "node_modules/domutils/lib/helpers.js"(exports2) {
    init_process_shim();
    exports2.removeSubsets = function(nodes) {
      var idx = nodes.length, node, ancestor, replace;
      while (--idx > -1) {
        node = ancestor = nodes[idx];
        nodes[idx] = null;
        replace = true;
        while (ancestor) {
          if (nodes.indexOf(ancestor) > -1) {
            replace = false;
            nodes.splice(idx, 1);
            break;
          }
          ancestor = ancestor.parent;
        }
        if (replace) {
          nodes[idx] = node;
        }
      }
      return nodes;
    };
    var POSITION = {
      DISCONNECTED: 1,
      PRECEDING: 2,
      FOLLOWING: 4,
      CONTAINS: 8,
      CONTAINED_BY: 16
    };
    var comparePos = exports2.compareDocumentPosition = function(nodeA, nodeB) {
      var aParents = [];
      var bParents = [];
      var current, sharedParent, siblings, aSibling, bSibling, idx;
      if (nodeA === nodeB) {
        return 0;
      }
      current = nodeA;
      while (current) {
        aParents.unshift(current);
        current = current.parent;
      }
      current = nodeB;
      while (current) {
        bParents.unshift(current);
        current = current.parent;
      }
      idx = 0;
      while (aParents[idx] === bParents[idx]) {
        idx++;
      }
      if (idx === 0) {
        return POSITION.DISCONNECTED;
      }
      sharedParent = aParents[idx - 1];
      siblings = sharedParent.children;
      aSibling = aParents[idx];
      bSibling = bParents[idx];
      if (siblings.indexOf(aSibling) > siblings.indexOf(bSibling)) {
        if (sharedParent === nodeB) {
          return POSITION.FOLLOWING | POSITION.CONTAINED_BY;
        }
        return POSITION.FOLLOWING;
      } else {
        if (sharedParent === nodeA) {
          return POSITION.PRECEDING | POSITION.CONTAINS;
        }
        return POSITION.PRECEDING;
      }
    };
    exports2.uniqueSort = function(nodes) {
      var idx = nodes.length, node, position;
      nodes = nodes.slice();
      while (--idx > -1) {
        node = nodes[idx];
        position = nodes.indexOf(node);
        if (position > -1 && position < idx) {
          nodes.splice(idx, 1);
        }
      }
      nodes.sort(function(a, b) {
        var relative = comparePos(a, b);
        if (relative & POSITION.PRECEDING) {
          return -1;
        } else if (relative & POSITION.FOLLOWING) {
          return 1;
        }
        return 0;
      });
      return nodes;
    };
  }
});

// node_modules/domutils/index.js
var require_domutils = __commonJS({
  "node_modules/domutils/index.js"(exports2, module2) {
    init_process_shim();
    var DomUtils = module2.exports;
    [
      require_stringify(),
      require_traversal(),
      require_manipulation(),
      require_querying(),
      require_legacy2(),
      require_helpers()
    ].forEach(function(ext) {
      Object.keys(ext).forEach(function(key) {
        DomUtils[key] = ext[key].bind(DomUtils);
      });
    });
  }
});

// node_modules/htmlparser2/lib/FeedHandler.js
var require_FeedHandler = __commonJS({
  "node_modules/htmlparser2/lib/FeedHandler.js"(exports2, module2) {
    init_process_shim();
    var DomHandler = require_domhandler();
    var DomUtils = require_domutils();
    function FeedHandler(callback, options) {
      this.init(callback, options);
    }
    require_inherits_browser()(FeedHandler, DomHandler);
    FeedHandler.prototype.init = DomHandler;
    function getElements(what, where) {
      return DomUtils.getElementsByTagName(what, where, true);
    }
    function getOneElement(what, where) {
      return DomUtils.getElementsByTagName(what, where, true, 1)[0];
    }
    function fetch(what, where, recurse) {
      return DomUtils.getText(
        DomUtils.getElementsByTagName(what, where, recurse, 1)
      ).trim();
    }
    function addConditionally(obj, prop, what, where, recurse) {
      var tmp = fetch(what, where, recurse);
      if (tmp) obj[prop] = tmp;
    }
    var isValidFeed = function(value) {
      return value === "rss" || value === "feed" || value === "rdf:RDF";
    };
    FeedHandler.prototype.onend = function() {
      var feed = {}, feedRoot = getOneElement(isValidFeed, this.dom), tmp, childs;
      if (feedRoot) {
        if (feedRoot.name === "feed") {
          childs = feedRoot.children;
          feed.type = "atom";
          addConditionally(feed, "id", "id", childs);
          addConditionally(feed, "title", "title", childs);
          if ((tmp = getOneElement("link", childs)) && (tmp = tmp.attribs) && (tmp = tmp.href))
            feed.link = tmp;
          addConditionally(feed, "description", "subtitle", childs);
          if (tmp = fetch("updated", childs)) feed.updated = new Date(tmp);
          addConditionally(feed, "author", "email", childs, true);
          feed.items = getElements("entry", childs).map(function(item) {
            var entry = {}, tmp2;
            item = item.children;
            addConditionally(entry, "id", "id", item);
            addConditionally(entry, "title", "title", item);
            if ((tmp2 = getOneElement("link", item)) && (tmp2 = tmp2.attribs) && (tmp2 = tmp2.href))
              entry.link = tmp2;
            if (tmp2 = fetch("summary", item) || fetch("content", item))
              entry.description = tmp2;
            if (tmp2 = fetch("updated", item))
              entry.pubDate = new Date(tmp2);
            return entry;
          });
        } else {
          childs = getOneElement("channel", feedRoot.children).children;
          feed.type = feedRoot.name.substr(0, 3);
          feed.id = "";
          addConditionally(feed, "title", "title", childs);
          addConditionally(feed, "link", "link", childs);
          addConditionally(feed, "description", "description", childs);
          if (tmp = fetch("lastBuildDate", childs))
            feed.updated = new Date(tmp);
          addConditionally(feed, "author", "managingEditor", childs, true);
          feed.items = getElements("item", feedRoot.children).map(function(item) {
            var entry = {}, tmp2;
            item = item.children;
            addConditionally(entry, "id", "guid", item);
            addConditionally(entry, "title", "title", item);
            addConditionally(entry, "link", "link", item);
            addConditionally(entry, "description", "description", item);
            if (tmp2 = fetch("pubDate", item))
              entry.pubDate = new Date(tmp2);
            return entry;
          });
        }
      }
      this.dom = feed;
      DomHandler.prototype._handleCallback.call(
        this,
        feedRoot ? null : Error("couldn't find root of feed")
      );
    };
    module2.exports = FeedHandler;
  }
});

// (disabled):node_modules/readable-stream/readable-browser.js
var require_readable_browser = __commonJS({
  "(disabled):node_modules/readable-stream/readable-browser.js"() {
    init_process_shim();
  }
});

// node_modules/htmlparser2/lib/WritableStream.js
var require_WritableStream = __commonJS({
  "node_modules/htmlparser2/lib/WritableStream.js"(exports2, module2) {
    init_process_shim();
    module2.exports = Stream;
    var Parser = require_Parser();
    var WritableStream = require_readable_browser().Writable;
    var StringDecoder = require("string_decoder").StringDecoder;
    var Buffer2 = require("buffer").Buffer;
    function Stream(cbs, options) {
      var parser = this._parser = new Parser(cbs, options);
      var decoder = this._decoder = new StringDecoder();
      WritableStream.call(this, { decodeStrings: false });
      this.once("finish", function() {
        parser.end(decoder.end());
      });
    }
    require_inherits_browser()(Stream, WritableStream);
    Stream.prototype._write = function(chunk, encoding, cb) {
      if (chunk instanceof Buffer2) chunk = this._decoder.write(chunk);
      this._parser.write(chunk);
      cb();
    };
  }
});

// node_modules/htmlparser2/lib/Stream.js
var require_Stream = __commonJS({
  "node_modules/htmlparser2/lib/Stream.js"(exports2, module2) {
    init_process_shim();
    module2.exports = Stream;
    var Parser = require_WritableStream();
    function Stream(options) {
      Parser.call(this, new Cbs(this), options);
    }
    require_inherits_browser()(Stream, Parser);
    Stream.prototype.readable = true;
    function Cbs(scope) {
      this.scope = scope;
    }
    var EVENTS = require_lib().EVENTS;
    Object.keys(EVENTS).forEach(function(name) {
      if (EVENTS[name] === 0) {
        Cbs.prototype["on" + name] = function() {
          this.scope.emit(name);
        };
      } else if (EVENTS[name] === 1) {
        Cbs.prototype["on" + name] = function(a) {
          this.scope.emit(name, a);
        };
      } else if (EVENTS[name] === 2) {
        Cbs.prototype["on" + name] = function(a, b) {
          this.scope.emit(name, a, b);
        };
      } else {
        throw Error("wrong number of arguments!");
      }
    });
  }
});

// node_modules/htmlparser2/lib/ProxyHandler.js
var require_ProxyHandler = __commonJS({
  "node_modules/htmlparser2/lib/ProxyHandler.js"(exports2, module2) {
    init_process_shim();
    module2.exports = ProxyHandler;
    function ProxyHandler(cbs) {
      this._cbs = cbs || {};
    }
    var EVENTS = require_lib().EVENTS;
    Object.keys(EVENTS).forEach(function(name) {
      if (EVENTS[name] === 0) {
        name = "on" + name;
        ProxyHandler.prototype[name] = function() {
          if (this._cbs[name]) this._cbs[name]();
        };
      } else if (EVENTS[name] === 1) {
        name = "on" + name;
        ProxyHandler.prototype[name] = function(a) {
          if (this._cbs[name]) this._cbs[name](a);
        };
      } else if (EVENTS[name] === 2) {
        name = "on" + name;
        ProxyHandler.prototype[name] = function(a, b) {
          if (this._cbs[name]) this._cbs[name](a, b);
        };
      } else {
        throw Error("wrong number of arguments");
      }
    });
  }
});

// node_modules/htmlparser2/lib/CollectingHandler.js
var require_CollectingHandler = __commonJS({
  "node_modules/htmlparser2/lib/CollectingHandler.js"(exports2, module2) {
    init_process_shim();
    module2.exports = CollectingHandler;
    function CollectingHandler(cbs) {
      this._cbs = cbs || {};
      this.events = [];
    }
    var EVENTS = require_lib().EVENTS;
    Object.keys(EVENTS).forEach(function(name) {
      if (EVENTS[name] === 0) {
        name = "on" + name;
        CollectingHandler.prototype[name] = function() {
          this.events.push([name]);
          if (this._cbs[name]) this._cbs[name]();
        };
      } else if (EVENTS[name] === 1) {
        name = "on" + name;
        CollectingHandler.prototype[name] = function(a) {
          this.events.push([name, a]);
          if (this._cbs[name]) this._cbs[name](a);
        };
      } else if (EVENTS[name] === 2) {
        name = "on" + name;
        CollectingHandler.prototype[name] = function(a, b) {
          this.events.push([name, a, b]);
          if (this._cbs[name]) this._cbs[name](a, b);
        };
      } else {
        throw Error("wrong number of arguments");
      }
    });
    CollectingHandler.prototype.onreset = function() {
      this.events = [];
      if (this._cbs.onreset) this._cbs.onreset();
    };
    CollectingHandler.prototype.restart = function() {
      if (this._cbs.onreset) this._cbs.onreset();
      for (var i = 0, len = this.events.length; i < len; i++) {
        if (this._cbs[this.events[i][0]]) {
          var num = this.events[i].length;
          if (num === 1) {
            this._cbs[this.events[i][0]]();
          } else if (num === 2) {
            this._cbs[this.events[i][0]](this.events[i][1]);
          } else {
            this._cbs[this.events[i][0]](
              this.events[i][1],
              this.events[i][2]
            );
          }
        }
      }
    };
  }
});

// node_modules/htmlparser2/lib/index.js
var require_lib = __commonJS({
  "node_modules/htmlparser2/lib/index.js"(exports2, module2) {
    init_process_shim();
    var Parser = require_Parser();
    var DomHandler = require_domhandler();
    function defineProp(name, value) {
      delete module2.exports[name];
      module2.exports[name] = value;
      return value;
    }
    module2.exports = {
      Parser,
      Tokenizer: require_Tokenizer(),
      ElementType: require_domelementtype(),
      DomHandler,
      get FeedHandler() {
        return defineProp("FeedHandler", require_FeedHandler());
      },
      get Stream() {
        return defineProp("Stream", require_Stream());
      },
      get WritableStream() {
        return defineProp("WritableStream", require_WritableStream());
      },
      get ProxyHandler() {
        return defineProp("ProxyHandler", require_ProxyHandler());
      },
      get DomUtils() {
        return defineProp("DomUtils", require_domutils());
      },
      get CollectingHandler() {
        return defineProp(
          "CollectingHandler",
          require_CollectingHandler()
        );
      },
      // For legacy support
      DefaultHandler: DomHandler,
      get RssHandler() {
        return defineProp("RssHandler", this.FeedHandler);
      },
      //helper methods
      parseDOM: function(data, options) {
        var handler = new DomHandler(options);
        new Parser(handler, options).end(data);
        return handler.dom;
      },
      parseFeed: function(feed, options) {
        var handler = new module2.exports.FeedHandler(options);
        new Parser(handler, options).end(feed);
        return handler.dom;
      },
      createDomStream: function(cb, options, elementCb) {
        var handler = new DomHandler(cb, options, elementCb);
        return new Parser(handler, options);
      },
      // List of all events that the parser emits
      EVENTS: {
        /* Format: eventname: number of arguments */
        attribute: 2,
        cdatastart: 0,
        cdataend: 0,
        text: 1,
        processinginstruction: 2,
        comment: 1,
        commentend: 0,
        closetag: 1,
        opentag: 2,
        opentagname: 1,
        error: 1,
        end: 0
      }
    };
  }
});

// node_modules/cheerio/lib/parse.js
var require_parse = __commonJS({
  "node_modules/cheerio/lib/parse.js"(exports2, module2) {
    init_process_shim();
    var htmlparser = require_lib();
    exports2 = module2.exports = function(content, options) {
      var dom = exports2.evaluate(content, options), root = exports2.evaluate("<root></root>", options)[0];
      root.type = "root";
      exports2.update(dom, root);
      return root;
    };
    exports2.evaluate = function(content, options) {
      var dom;
      if (typeof content === "string" || Buffer.isBuffer(content)) {
        dom = htmlparser.parseDOM(content, options);
      } else {
        dom = content;
      }
      return dom;
    };
    exports2.update = function(arr, parent) {
      if (!Array.isArray(arr)) arr = [arr];
      if (parent) {
        parent.children = arr;
      } else {
        parent = null;
      }
      for (var i = 0; i < arr.length; i++) {
        var node = arr[i];
        var oldParent = node.parent || node.root, oldSiblings = oldParent && oldParent.children;
        if (oldSiblings && oldSiblings !== arr) {
          oldSiblings.splice(oldSiblings.indexOf(node), 1);
          if (node.prev) {
            node.prev.next = node.next;
          }
          if (node.next) {
            node.next.prev = node.prev;
          }
        }
        if (parent) {
          node.prev = arr[i - 1] || null;
          node.next = arr[i + 1] || null;
        } else {
          node.prev = node.next = null;
        }
        if (parent && parent.type === "root") {
          node.root = parent;
          node.parent = null;
        } else {
          node.root = null;
          node.parent = parent;
        }
      }
      return parent;
    };
  }
});

// node_modules/cheerio/lib/utils.js
var require_utils = __commonJS({
  "node_modules/cheerio/lib/utils.js"(exports2) {
    init_process_shim();
    var parse = require_parse();
    var render = require_dom_serializer();
    var tags = { tag: true, script: true, style: true };
    exports2.isTag = function(type) {
      if (type.type) type = type.type;
      return tags[type] || false;
    };
    exports2.camelCase = function(str) {
      return str.replace(/[_.-](\w|$)/g, function(_, x) {
        return x.toUpperCase();
      });
    };
    exports2.cssCase = function(str) {
      return str.replace(/[A-Z]/g, "-$&").toLowerCase();
    };
    exports2.domEach = function(cheerio, fn) {
      var i = 0, len = cheerio.length;
      while (i < len && fn.call(cheerio, i, cheerio[i]) !== false) ++i;
      return cheerio;
    };
    exports2.cloneDom = function(dom, options) {
      return parse(render(dom, options), options).children;
    };
    var quickExpr = /^(?:[^#<]*(<[\w\W]+>)[^>]*$|#([\w\-]*)$)/;
    exports2.isHtml = function(str) {
      if (str.charAt(0) === "<" && str.charAt(str.length - 1) === ">" && str.length >= 3) return true;
      var match = quickExpr.exec(str);
      return !!(match && match[1]);
    };
  }
});

// node_modules/lodash.assignin/index.js
var require_lodash = __commonJS({
  "node_modules/lodash.assignin/index.js"(exports2, module2) {
    init_process_shim();
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    function apply(func, thisArg, args) {
      switch (args.length) {
        case 0:
          return func.call(thisArg);
        case 1:
          return func.call(thisArg, args[0]);
        case 2:
          return func.call(thisArg, args[0], args[1]);
        case 3:
          return func.call(thisArg, args[0], args[1], args[2]);
      }
      return func.apply(thisArg, args);
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    var objectProto = Object.prototype;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var nativeMax = Math.max;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assignValue(object, key, value) {
      var objValue = object[key];
      if (!(hasOwnProperty2.call(object, key) && eq(objValue, value)) || value === void 0 && !(key in object)) {
        object[key] = value;
      }
    }
    function baseKeysIn(object) {
      if (!isObject(object)) {
        return nativeKeysIn(object);
      }
      var isProto = isPrototype(object), result = [];
      for (var key in object) {
        if (!(key == "constructor" && (isProto || !hasOwnProperty2.call(object, key)))) {
          result.push(key);
        }
      }
      return result;
    }
    function baseRest(func, start) {
      start = nativeMax(start === void 0 ? func.length - 1 : start, 0);
      return function() {
        var args = arguments, index = -1, length = nativeMax(args.length - start, 0), array = Array(length);
        while (++index < length) {
          array[index] = args[start + index];
        }
        index = -1;
        var otherArgs = Array(start + 1);
        while (++index < start) {
          otherArgs[index] = args[index];
        }
        otherArgs[start] = array;
        return apply(func, this, otherArgs);
      };
    }
    function copyObject(source, props, object, customizer) {
      object || (object = {});
      var index = -1, length = props.length;
      while (++index < length) {
        var key = props[index];
        var newValue = customizer ? customizer(object[key], source[key], key, object, source) : void 0;
        assignValue(object, key, newValue === void 0 ? source[key] : newValue);
      }
      return object;
    }
    function createAssigner(assigner) {
      return baseRest(function(object, sources) {
        var index = -1, length = sources.length, customizer = length > 1 ? sources[length - 1] : void 0, guard = length > 2 ? sources[2] : void 0;
        customizer = assigner.length > 3 && typeof customizer == "function" ? (length--, customizer) : void 0;
        if (guard && isIterateeCall(sources[0], sources[1], guard)) {
          customizer = length < 3 ? void 0 : customizer;
          length = 1;
        }
        object = Object(object);
        while (++index < length) {
          var source = sources[index];
          if (source) {
            assigner(object, source, index, customizer);
          }
        }
        return object;
      });
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isIterateeCall(value, index, object) {
      if (!isObject(object)) {
        return false;
      }
      var type = typeof index;
      if (type == "number" ? isArrayLike(object) && isIndex(index, object.length) : type == "string" && index in object) {
        return eq(object[index], value);
      }
      return false;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function nativeKeysIn(object) {
      var result = [];
      if (object != null) {
        for (var key in Object(object)) {
          result.push(key);
        }
      }
      return result;
    }
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    var assignIn = createAssigner(function(object, source) {
      copyObject(source, keysIn(source), object);
    });
    function keysIn(object) {
      return isArrayLike(object) ? arrayLikeKeys(object, true) : baseKeysIn(object);
    }
    module2.exports = assignIn;
  }
});

// node_modules/lodash.bind/index.js
var require_lodash2 = __commonJS({
  "node_modules/lodash.bind/index.js"(exports2, module2) {
    init_process_shim();
    var FUNC_ERROR_TEXT = "Expected a function";
    var PLACEHOLDER = "__lodash_placeholder__";
    var BIND_FLAG = 1;
    var BIND_KEY_FLAG = 2;
    var CURRY_BOUND_FLAG = 4;
    var CURRY_FLAG = 8;
    var CURRY_RIGHT_FLAG = 16;
    var PARTIAL_FLAG = 32;
    var PARTIAL_RIGHT_FLAG = 64;
    var ARY_FLAG = 128;
    var REARG_FLAG = 256;
    var FLIP_FLAG = 512;
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var MAX_INTEGER = 17976931348623157e292;
    var NAN = 0 / 0;
    var wrapFlags = [
      ["ary", ARY_FLAG],
      ["bind", BIND_FLAG],
      ["bindKey", BIND_KEY_FLAG],
      ["curry", CURRY_FLAG],
      ["curryRight", CURRY_RIGHT_FLAG],
      ["flip", FLIP_FLAG],
      ["partial", PARTIAL_FLAG],
      ["partialRight", PARTIAL_RIGHT_FLAG],
      ["rearg", REARG_FLAG]
    ];
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var symbolTag = "[object Symbol]";
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reTrim = /^\s+|\s+$/g;
    var reWrapComment = /\{(?:\n\/\* \[wrapped with .+\] \*\/)?\n?/;
    var reWrapDetails = /\{\n\/\* \[wrapped with (.+)\] \*/;
    var reSplitDetails = /,? & /;
    var reIsBadHex = /^[-+]0x[0-9a-f]+$/i;
    var reIsBinary = /^0b[01]+$/i;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsOctal = /^0o[0-7]+$/i;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var freeParseInt = parseInt;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    function apply(func, thisArg, args) {
      switch (args.length) {
        case 0:
          return func.call(thisArg);
        case 1:
          return func.call(thisArg, args[0]);
        case 2:
          return func.call(thisArg, args[0], args[1]);
        case 3:
          return func.call(thisArg, args[0], args[1], args[2]);
      }
      return func.apply(thisArg, args);
    }
    function arrayEach(array, iteratee) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (iteratee(array[index], index, array) === false) {
          break;
        }
      }
      return array;
    }
    function arrayIncludes(array, value) {
      var length = array ? array.length : 0;
      return !!length && baseIndexOf(array, value, 0) > -1;
    }
    function baseFindIndex(array, predicate, fromIndex, fromRight) {
      var length = array.length, index = fromIndex + (fromRight ? 1 : -1);
      while (fromRight ? index-- : ++index < length) {
        if (predicate(array[index], index, array)) {
          return index;
        }
      }
      return -1;
    }
    function baseIndexOf(array, value, fromIndex) {
      if (value !== value) {
        return baseFindIndex(array, baseIsNaN, fromIndex);
      }
      var index = fromIndex - 1, length = array.length;
      while (++index < length) {
        if (array[index] === value) {
          return index;
        }
      }
      return -1;
    }
    function baseIsNaN(value) {
      return value !== value;
    }
    function countHolders(array, placeholder) {
      var length = array.length, result = 0;
      while (length--) {
        if (array[length] === placeholder) {
          result++;
        }
      }
      return result;
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function isHostObject(value) {
      var result = false;
      if (value != null && typeof value.toString != "function") {
        try {
          result = !!(value + "");
        } catch (e) {
        }
      }
      return result;
    }
    function replaceHolders(array, placeholder) {
      var index = -1, length = array.length, resIndex = 0, result = [];
      while (++index < length) {
        var value = array[index];
        if (value === placeholder || value === PLACEHOLDER) {
          array[index] = PLACEHOLDER;
          result[resIndex++] = index;
        }
      }
      return result;
    }
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var objectCreate = Object.create;
    var nativeMax = Math.max;
    var nativeMin = Math.min;
    var defineProperty = (function() {
      var func = getNative(Object, "defineProperty"), name = getNative.name;
      return name && name.length > 2 ? func : void 0;
    })();
    function baseCreate(proto) {
      return isObject(proto) ? objectCreate(proto) : {};
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) || isHostObject(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseRest(func, start) {
      start = nativeMax(start === void 0 ? func.length - 1 : start, 0);
      return function() {
        var args = arguments, index = -1, length = nativeMax(args.length - start, 0), array = Array(length);
        while (++index < length) {
          array[index] = args[start + index];
        }
        index = -1;
        var otherArgs = Array(start + 1);
        while (++index < start) {
          otherArgs[index] = args[index];
        }
        otherArgs[start] = array;
        return apply(func, this, otherArgs);
      };
    }
    function composeArgs(args, partials, holders, isCurried) {
      var argsIndex = -1, argsLength = args.length, holdersLength = holders.length, leftIndex = -1, leftLength = partials.length, rangeLength = nativeMax(argsLength - holdersLength, 0), result = Array(leftLength + rangeLength), isUncurried = !isCurried;
      while (++leftIndex < leftLength) {
        result[leftIndex] = partials[leftIndex];
      }
      while (++argsIndex < holdersLength) {
        if (isUncurried || argsIndex < argsLength) {
          result[holders[argsIndex]] = args[argsIndex];
        }
      }
      while (rangeLength--) {
        result[leftIndex++] = args[argsIndex++];
      }
      return result;
    }
    function composeArgsRight(args, partials, holders, isCurried) {
      var argsIndex = -1, argsLength = args.length, holdersIndex = -1, holdersLength = holders.length, rightIndex = -1, rightLength = partials.length, rangeLength = nativeMax(argsLength - holdersLength, 0), result = Array(rangeLength + rightLength), isUncurried = !isCurried;
      while (++argsIndex < rangeLength) {
        result[argsIndex] = args[argsIndex];
      }
      var offset = argsIndex;
      while (++rightIndex < rightLength) {
        result[offset + rightIndex] = partials[rightIndex];
      }
      while (++holdersIndex < holdersLength) {
        if (isUncurried || argsIndex < argsLength) {
          result[offset + holders[holdersIndex]] = args[argsIndex++];
        }
      }
      return result;
    }
    function copyArray(source, array) {
      var index = -1, length = source.length;
      array || (array = Array(length));
      while (++index < length) {
        array[index] = source[index];
      }
      return array;
    }
    function createBind(func, bitmask, thisArg) {
      var isBind = bitmask & BIND_FLAG, Ctor = createCtor(func);
      function wrapper() {
        var fn = this && this !== root && this instanceof wrapper ? Ctor : func;
        return fn.apply(isBind ? thisArg : this, arguments);
      }
      return wrapper;
    }
    function createCtor(Ctor) {
      return function() {
        var args = arguments;
        switch (args.length) {
          case 0:
            return new Ctor();
          case 1:
            return new Ctor(args[0]);
          case 2:
            return new Ctor(args[0], args[1]);
          case 3:
            return new Ctor(args[0], args[1], args[2]);
          case 4:
            return new Ctor(args[0], args[1], args[2], args[3]);
          case 5:
            return new Ctor(args[0], args[1], args[2], args[3], args[4]);
          case 6:
            return new Ctor(args[0], args[1], args[2], args[3], args[4], args[5]);
          case 7:
            return new Ctor(args[0], args[1], args[2], args[3], args[4], args[5], args[6]);
        }
        var thisBinding = baseCreate(Ctor.prototype), result = Ctor.apply(thisBinding, args);
        return isObject(result) ? result : thisBinding;
      };
    }
    function createCurry(func, bitmask, arity) {
      var Ctor = createCtor(func);
      function wrapper() {
        var length = arguments.length, args = Array(length), index = length, placeholder = getHolder(wrapper);
        while (index--) {
          args[index] = arguments[index];
        }
        var holders = length < 3 && args[0] !== placeholder && args[length - 1] !== placeholder ? [] : replaceHolders(args, placeholder);
        length -= holders.length;
        if (length < arity) {
          return createRecurry(
            func,
            bitmask,
            createHybrid,
            wrapper.placeholder,
            void 0,
            args,
            holders,
            void 0,
            void 0,
            arity - length
          );
        }
        var fn = this && this !== root && this instanceof wrapper ? Ctor : func;
        return apply(fn, this, args);
      }
      return wrapper;
    }
    function createHybrid(func, bitmask, thisArg, partials, holders, partialsRight, holdersRight, argPos, ary, arity) {
      var isAry = bitmask & ARY_FLAG, isBind = bitmask & BIND_FLAG, isBindKey = bitmask & BIND_KEY_FLAG, isCurried = bitmask & (CURRY_FLAG | CURRY_RIGHT_FLAG), isFlip = bitmask & FLIP_FLAG, Ctor = isBindKey ? void 0 : createCtor(func);
      function wrapper() {
        var length = arguments.length, args = Array(length), index = length;
        while (index--) {
          args[index] = arguments[index];
        }
        if (isCurried) {
          var placeholder = getHolder(wrapper), holdersCount = countHolders(args, placeholder);
        }
        if (partials) {
          args = composeArgs(args, partials, holders, isCurried);
        }
        if (partialsRight) {
          args = composeArgsRight(args, partialsRight, holdersRight, isCurried);
        }
        length -= holdersCount;
        if (isCurried && length < arity) {
          var newHolders = replaceHolders(args, placeholder);
          return createRecurry(
            func,
            bitmask,
            createHybrid,
            wrapper.placeholder,
            thisArg,
            args,
            newHolders,
            argPos,
            ary,
            arity - length
          );
        }
        var thisBinding = isBind ? thisArg : this, fn = isBindKey ? thisBinding[func] : func;
        length = args.length;
        if (argPos) {
          args = reorder(args, argPos);
        } else if (isFlip && length > 1) {
          args.reverse();
        }
        if (isAry && ary < length) {
          args.length = ary;
        }
        if (this && this !== root && this instanceof wrapper) {
          fn = Ctor || createCtor(fn);
        }
        return fn.apply(thisBinding, args);
      }
      return wrapper;
    }
    function createPartial(func, bitmask, thisArg, partials) {
      var isBind = bitmask & BIND_FLAG, Ctor = createCtor(func);
      function wrapper() {
        var argsIndex = -1, argsLength = arguments.length, leftIndex = -1, leftLength = partials.length, args = Array(leftLength + argsLength), fn = this && this !== root && this instanceof wrapper ? Ctor : func;
        while (++leftIndex < leftLength) {
          args[leftIndex] = partials[leftIndex];
        }
        while (argsLength--) {
          args[leftIndex++] = arguments[++argsIndex];
        }
        return apply(fn, isBind ? thisArg : this, args);
      }
      return wrapper;
    }
    function createRecurry(func, bitmask, wrapFunc, placeholder, thisArg, partials, holders, argPos, ary, arity) {
      var isCurry = bitmask & CURRY_FLAG, newHolders = isCurry ? holders : void 0, newHoldersRight = isCurry ? void 0 : holders, newPartials = isCurry ? partials : void 0, newPartialsRight = isCurry ? void 0 : partials;
      bitmask |= isCurry ? PARTIAL_FLAG : PARTIAL_RIGHT_FLAG;
      bitmask &= ~(isCurry ? PARTIAL_RIGHT_FLAG : PARTIAL_FLAG);
      if (!(bitmask & CURRY_BOUND_FLAG)) {
        bitmask &= ~(BIND_FLAG | BIND_KEY_FLAG);
      }
      var result = wrapFunc(func, bitmask, thisArg, newPartials, newHolders, newPartialsRight, newHoldersRight, argPos, ary, arity);
      result.placeholder = placeholder;
      return setWrapToString(result, func, bitmask);
    }
    function createWrap(func, bitmask, thisArg, partials, holders, argPos, ary, arity) {
      var isBindKey = bitmask & BIND_KEY_FLAG;
      if (!isBindKey && typeof func != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      var length = partials ? partials.length : 0;
      if (!length) {
        bitmask &= ~(PARTIAL_FLAG | PARTIAL_RIGHT_FLAG);
        partials = holders = void 0;
      }
      ary = ary === void 0 ? ary : nativeMax(toInteger(ary), 0);
      arity = arity === void 0 ? arity : toInteger(arity);
      length -= holders ? holders.length : 0;
      if (bitmask & PARTIAL_RIGHT_FLAG) {
        var partialsRight = partials, holdersRight = holders;
        partials = holders = void 0;
      }
      var newData = [
        func,
        bitmask,
        thisArg,
        partials,
        holders,
        partialsRight,
        holdersRight,
        argPos,
        ary,
        arity
      ];
      func = newData[0];
      bitmask = newData[1];
      thisArg = newData[2];
      partials = newData[3];
      holders = newData[4];
      arity = newData[9] = newData[9] == null ? isBindKey ? 0 : func.length : nativeMax(newData[9] - length, 0);
      if (!arity && bitmask & (CURRY_FLAG | CURRY_RIGHT_FLAG)) {
        bitmask &= ~(CURRY_FLAG | CURRY_RIGHT_FLAG);
      }
      if (!bitmask || bitmask == BIND_FLAG) {
        var result = createBind(func, bitmask, thisArg);
      } else if (bitmask == CURRY_FLAG || bitmask == CURRY_RIGHT_FLAG) {
        result = createCurry(func, bitmask, arity);
      } else if ((bitmask == PARTIAL_FLAG || bitmask == (BIND_FLAG | PARTIAL_FLAG)) && !holders.length) {
        result = createPartial(func, bitmask, thisArg, partials);
      } else {
        result = createHybrid.apply(void 0, newData);
      }
      return setWrapToString(result, func, bitmask);
    }
    function getHolder(func) {
      var object = func;
      return object.placeholder;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    function getWrapDetails(source) {
      var match = source.match(reWrapDetails);
      return match ? match[1].split(reSplitDetails) : [];
    }
    function insertWrapDetails(source, details) {
      var length = details.length, lastIndex = length - 1;
      details[lastIndex] = (length > 1 ? "& " : "") + details[lastIndex];
      details = details.join(length > 2 ? ", " : " ");
      return source.replace(reWrapComment, "{\n/* [wrapped with " + details + "] */\n");
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function reorder(array, indexes) {
      var arrLength = array.length, length = nativeMin(indexes.length, arrLength), oldArray = copyArray(array);
      while (length--) {
        var index = indexes[length];
        array[length] = isIndex(index, arrLength) ? oldArray[index] : void 0;
      }
      return array;
    }
    var setWrapToString = !defineProperty ? identity : function(wrapper, reference, bitmask) {
      var source = reference + "";
      return defineProperty(wrapper, "toString", {
        "configurable": true,
        "enumerable": false,
        "value": constant(insertWrapDetails(source, updateWrapDetails(getWrapDetails(source), bitmask)))
      });
    };
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function updateWrapDetails(details, bitmask) {
      arrayEach(wrapFlags, function(pair) {
        var value = "_." + pair[0];
        if (bitmask & pair[1] && !arrayIncludes(details, value)) {
          details.push(value);
        }
      });
      return details.sort();
    }
    var bind = baseRest(function(func, thisArg, partials) {
      var bitmask = BIND_FLAG;
      if (partials.length) {
        var holders = replaceHolders(partials, getHolder(bind));
        bitmask |= PARTIAL_FLAG;
      }
      return createWrap(func, bitmask, thisArg, partials, holders);
    });
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    function toFinite(value) {
      if (!value) {
        return value === 0 ? value : 0;
      }
      value = toNumber(value);
      if (value === INFINITY || value === -INFINITY) {
        var sign = value < 0 ? -1 : 1;
        return sign * MAX_INTEGER;
      }
      return value === value ? value : 0;
    }
    function toInteger(value) {
      var result = toFinite(value), remainder = result % 1;
      return result === result ? remainder ? result - remainder : result : 0;
    }
    function toNumber(value) {
      if (typeof value == "number") {
        return value;
      }
      if (isSymbol(value)) {
        return NAN;
      }
      if (isObject(value)) {
        var other = typeof value.valueOf == "function" ? value.valueOf() : value;
        value = isObject(other) ? other + "" : other;
      }
      if (typeof value != "string") {
        return value === 0 ? value : +value;
      }
      value = value.replace(reTrim, "");
      var isBinary = reIsBinary.test(value);
      return isBinary || reIsOctal.test(value) ? freeParseInt(value.slice(2), isBinary ? 2 : 8) : reIsBadHex.test(value) ? NAN : +value;
    }
    function constant(value) {
      return function() {
        return value;
      };
    }
    function identity(value) {
      return value;
    }
    bind.placeholder = {};
    module2.exports = bind;
  }
});

// node_modules/lodash.foreach/index.js
var require_lodash3 = __commonJS({
  "node_modules/lodash.foreach/index.js"(exports2, module2) {
    init_process_shim();
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    function arrayEach(array, iteratee) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (iteratee(array[index], index, array) === false) {
          break;
        }
      }
      return array;
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    var objectProto = Object.prototype;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var nativeKeys = overArg(Object.keys, Object);
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    var baseEach = createBaseEach(baseForOwn);
    var baseFor = createBaseFor();
    function baseForOwn(object, iteratee) {
      return object && baseFor(object, iteratee, keys);
    }
    function baseKeys(object) {
      if (!isPrototype(object)) {
        return nativeKeys(object);
      }
      var result = [];
      for (var key in Object(object)) {
        if (hasOwnProperty2.call(object, key) && key != "constructor") {
          result.push(key);
        }
      }
      return result;
    }
    function createBaseEach(eachFunc, fromRight) {
      return function(collection, iteratee) {
        if (collection == null) {
          return collection;
        }
        if (!isArrayLike(collection)) {
          return eachFunc(collection, iteratee);
        }
        var length = collection.length, index = fromRight ? length : -1, iterable = Object(collection);
        while (fromRight ? index-- : ++index < length) {
          if (iteratee(iterable[index], index, iterable) === false) {
            break;
          }
        }
        return collection;
      };
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function forEach(collection, iteratee) {
      var func = isArray(collection) ? arrayEach : baseEach;
      return func(collection, typeof iteratee == "function" ? iteratee : identity);
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function keys(object) {
      return isArrayLike(object) ? arrayLikeKeys(object) : baseKeys(object);
    }
    function identity(value) {
      return value;
    }
    module2.exports = forEach;
  }
});

// node_modules/lodash.defaults/index.js
var require_lodash4 = __commonJS({
  "node_modules/lodash.defaults/index.js"(exports2, module2) {
    init_process_shim();
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    function apply(func, thisArg, args) {
      switch (args.length) {
        case 0:
          return func.call(thisArg);
        case 1:
          return func.call(thisArg, args[0]);
        case 2:
          return func.call(thisArg, args[0], args[1]);
        case 3:
          return func.call(thisArg, args[0], args[1], args[2]);
      }
      return func.apply(thisArg, args);
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    var objectProto = Object.prototype;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var nativeMax = Math.max;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assignInDefaults(objValue, srcValue, key, object) {
      if (objValue === void 0 || eq(objValue, objectProto[key]) && !hasOwnProperty2.call(object, key)) {
        return srcValue;
      }
      return objValue;
    }
    function assignValue(object, key, value) {
      var objValue = object[key];
      if (!(hasOwnProperty2.call(object, key) && eq(objValue, value)) || value === void 0 && !(key in object)) {
        object[key] = value;
      }
    }
    function baseKeysIn(object) {
      if (!isObject(object)) {
        return nativeKeysIn(object);
      }
      var isProto = isPrototype(object), result = [];
      for (var key in object) {
        if (!(key == "constructor" && (isProto || !hasOwnProperty2.call(object, key)))) {
          result.push(key);
        }
      }
      return result;
    }
    function baseRest(func, start) {
      start = nativeMax(start === void 0 ? func.length - 1 : start, 0);
      return function() {
        var args = arguments, index = -1, length = nativeMax(args.length - start, 0), array = Array(length);
        while (++index < length) {
          array[index] = args[start + index];
        }
        index = -1;
        var otherArgs = Array(start + 1);
        while (++index < start) {
          otherArgs[index] = args[index];
        }
        otherArgs[start] = array;
        return apply(func, this, otherArgs);
      };
    }
    function copyObject(source, props, object, customizer) {
      object || (object = {});
      var index = -1, length = props.length;
      while (++index < length) {
        var key = props[index];
        var newValue = customizer ? customizer(object[key], source[key], key, object, source) : void 0;
        assignValue(object, key, newValue === void 0 ? source[key] : newValue);
      }
      return object;
    }
    function createAssigner(assigner) {
      return baseRest(function(object, sources) {
        var index = -1, length = sources.length, customizer = length > 1 ? sources[length - 1] : void 0, guard = length > 2 ? sources[2] : void 0;
        customizer = assigner.length > 3 && typeof customizer == "function" ? (length--, customizer) : void 0;
        if (guard && isIterateeCall(sources[0], sources[1], guard)) {
          customizer = length < 3 ? void 0 : customizer;
          length = 1;
        }
        object = Object(object);
        while (++index < length) {
          var source = sources[index];
          if (source) {
            assigner(object, source, index, customizer);
          }
        }
        return object;
      });
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isIterateeCall(value, index, object) {
      if (!isObject(object)) {
        return false;
      }
      var type = typeof index;
      if (type == "number" ? isArrayLike(object) && isIndex(index, object.length) : type == "string" && index in object) {
        return eq(object[index], value);
      }
      return false;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function nativeKeysIn(object) {
      var result = [];
      if (object != null) {
        for (var key in Object(object)) {
          result.push(key);
        }
      }
      return result;
    }
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    var assignInWith = createAssigner(function(object, source, srcIndex, customizer) {
      copyObject(source, keysIn(source), object, customizer);
    });
    var defaults = baseRest(function(args) {
      args.push(void 0, assignInDefaults);
      return apply(assignInWith, void 0, args);
    });
    function keysIn(object) {
      return isArrayLike(object) ? arrayLikeKeys(object, true) : baseKeysIn(object);
    }
    module2.exports = defaults;
  }
});

// node_modules/nth-check/parse.js
var require_parse2 = __commonJS({
  "node_modules/nth-check/parse.js"(exports2, module2) {
    init_process_shim();
    module2.exports = parse;
    var re_nthElement = /^([+\-]?\d*n)?\s*(?:([+\-]?)\s*(\d+))?$/;
    function parse(formula) {
      formula = formula.trim().toLowerCase();
      if (formula === "even") {
        return [2, 0];
      } else if (formula === "odd") {
        return [2, 1];
      } else {
        var parsed = formula.match(re_nthElement);
        if (!parsed) {
          throw new SyntaxError("n-th rule couldn't be parsed ('" + formula + "')");
        }
        var a;
        if (parsed[1]) {
          a = parseInt(parsed[1], 10);
          if (isNaN(a)) {
            if (parsed[1].charAt(0) === "-") a = -1;
            else a = 1;
          }
        } else a = 0;
        return [
          a,
          parsed[3] ? parseInt((parsed[2] || "") + parsed[3], 10) : 0
        ];
      }
    }
  }
});

// node_modules/boolbase/index.js
var require_boolbase = __commonJS({
  "node_modules/boolbase/index.js"(exports2, module2) {
    init_process_shim();
    module2.exports = {
      trueFunc: function trueFunc() {
        return true;
      },
      falseFunc: function falseFunc() {
        return false;
      }
    };
  }
});

// node_modules/nth-check/compile.js
var require_compile = __commonJS({
  "node_modules/nth-check/compile.js"(exports2, module2) {
    init_process_shim();
    module2.exports = compile;
    var BaseFuncs = require_boolbase();
    var trueFunc = BaseFuncs.trueFunc;
    var falseFunc = BaseFuncs.falseFunc;
    function compile(parsed) {
      var a = parsed[0], b = parsed[1] - 1;
      if (b < 0 && a <= 0) return falseFunc;
      if (a === -1) return function(pos) {
        return pos <= b;
      };
      if (a === 0) return function(pos) {
        return pos === b;
      };
      if (a === 1) return b < 0 ? trueFunc : function(pos) {
        return pos >= b;
      };
      var bMod = b % a;
      if (bMod < 0) bMod += a;
      if (a > 1) {
        return function(pos) {
          return pos >= b && pos % a === bMod;
        };
      }
      a *= -1;
      return function(pos) {
        return pos <= b && pos % a === bMod;
      };
    }
  }
});

// node_modules/nth-check/index.js
var require_nth_check = __commonJS({
  "node_modules/nth-check/index.js"(exports2, module2) {
    init_process_shim();
    var parse = require_parse2();
    var compile = require_compile();
    module2.exports = function nthCheck(formula) {
      return compile(parse(formula));
    };
    module2.exports.parse = parse;
    module2.exports.compile = compile;
  }
});

// node_modules/css-select/lib/attributes.js
var require_attributes = __commonJS({
  "node_modules/css-select/lib/attributes.js"(exports2, module2) {
    init_process_shim();
    var DomUtils = require_domutils();
    var hasAttrib = DomUtils.hasAttrib;
    var getAttributeValue = DomUtils.getAttributeValue;
    var falseFunc = require_boolbase().falseFunc;
    var reChars = /[-[\]{}()*+?.,\\^$|#\s]/g;
    var attributeRules = {
      __proto__: null,
      equals: function(next, data) {
        var name = data.name, value = data.value;
        if (data.ignoreCase) {
          value = value.toLowerCase();
          return function equalsIC(elem) {
            var attr = getAttributeValue(elem, name);
            return attr != null && attr.toLowerCase() === value && next(elem);
          };
        }
        return function equals(elem) {
          return getAttributeValue(elem, name) === value && next(elem);
        };
      },
      hyphen: function(next, data) {
        var name = data.name, value = data.value, len = value.length;
        if (data.ignoreCase) {
          value = value.toLowerCase();
          return function hyphenIC(elem) {
            var attr = getAttributeValue(elem, name);
            return attr != null && (attr.length === len || attr.charAt(len) === "-") && attr.substr(0, len).toLowerCase() === value && next(elem);
          };
        }
        return function hyphen(elem) {
          var attr = getAttributeValue(elem, name);
          return attr != null && attr.substr(0, len) === value && (attr.length === len || attr.charAt(len) === "-") && next(elem);
        };
      },
      element: function(next, data) {
        var name = data.name, value = data.value;
        if (/\s/.test(value)) {
          return falseFunc;
        }
        value = value.replace(reChars, "\\$&");
        var pattern = "(?:^|\\s)" + value + "(?:$|\\s)", flags = data.ignoreCase ? "i" : "", regex = new RegExp(pattern, flags);
        return function element(elem) {
          var attr = getAttributeValue(elem, name);
          return attr != null && regex.test(attr) && next(elem);
        };
      },
      exists: function(next, data) {
        var name = data.name;
        return function exists(elem) {
          return hasAttrib(elem, name) && next(elem);
        };
      },
      start: function(next, data) {
        var name = data.name, value = data.value, len = value.length;
        if (len === 0) {
          return falseFunc;
        }
        if (data.ignoreCase) {
          value = value.toLowerCase();
          return function startIC(elem) {
            var attr = getAttributeValue(elem, name);
            return attr != null && attr.substr(0, len).toLowerCase() === value && next(elem);
          };
        }
        return function start(elem) {
          var attr = getAttributeValue(elem, name);
          return attr != null && attr.substr(0, len) === value && next(elem);
        };
      },
      end: function(next, data) {
        var name = data.name, value = data.value, len = -value.length;
        if (len === 0) {
          return falseFunc;
        }
        if (data.ignoreCase) {
          value = value.toLowerCase();
          return function endIC(elem) {
            var attr = getAttributeValue(elem, name);
            return attr != null && attr.substr(len).toLowerCase() === value && next(elem);
          };
        }
        return function end(elem) {
          var attr = getAttributeValue(elem, name);
          return attr != null && attr.substr(len) === value && next(elem);
        };
      },
      any: function(next, data) {
        var name = data.name, value = data.value;
        if (value === "") {
          return falseFunc;
        }
        if (data.ignoreCase) {
          var regex = new RegExp(value.replace(reChars, "\\$&"), "i");
          return function anyIC(elem) {
            var attr = getAttributeValue(elem, name);
            return attr != null && regex.test(attr) && next(elem);
          };
        }
        return function any(elem) {
          var attr = getAttributeValue(elem, name);
          return attr != null && attr.indexOf(value) >= 0 && next(elem);
        };
      },
      not: function(next, data) {
        var name = data.name, value = data.value;
        if (value === "") {
          return function notEmpty(elem) {
            return !!getAttributeValue(elem, name) && next(elem);
          };
        } else if (data.ignoreCase) {
          value = value.toLowerCase();
          return function notIC(elem) {
            var attr = getAttributeValue(elem, name);
            return attr != null && attr.toLowerCase() !== value && next(elem);
          };
        }
        return function not(elem) {
          return getAttributeValue(elem, name) !== value && next(elem);
        };
      }
    };
    module2.exports = {
      compile: function(next, data, options) {
        if (options && options.strict && (data.ignoreCase || data.action === "not")) throw SyntaxError("Unsupported attribute selector");
        return attributeRules[data.action](next, data);
      },
      rules: attributeRules
    };
  }
});

// node_modules/css-select/lib/pseudos.js
var require_pseudos = __commonJS({
  "node_modules/css-select/lib/pseudos.js"(exports2, module2) {
    init_process_shim();
    var DomUtils = require_domutils();
    var isTag = DomUtils.isTag;
    var getText = DomUtils.getText;
    var getParent = DomUtils.getParent;
    var getChildren = DomUtils.getChildren;
    var getSiblings = DomUtils.getSiblings;
    var hasAttrib = DomUtils.hasAttrib;
    var getName = DomUtils.getName;
    var getAttribute = DomUtils.getAttributeValue;
    var getNCheck = require_nth_check();
    var checkAttrib = require_attributes().rules.equals;
    var BaseFuncs = require_boolbase();
    var trueFunc = BaseFuncs.trueFunc;
    var falseFunc = BaseFuncs.falseFunc;
    function getFirstElement(elems) {
      for (var i = 0; elems && i < elems.length; i++) {
        if (isTag(elems[i])) return elems[i];
      }
    }
    function getAttribFunc(name, value) {
      var data = { name, value };
      return function attribFunc(next) {
        return checkAttrib(next, data);
      };
    }
    function getChildFunc(next) {
      return function(elem) {
        return !!getParent(elem) && next(elem);
      };
    }
    var filters = {
      contains: function(next, text) {
        return function contains(elem) {
          return next(elem) && getText(elem).indexOf(text) >= 0;
        };
      },
      icontains: function(next, text) {
        var itext = text.toLowerCase();
        return function icontains(elem) {
          return next(elem) && getText(elem).toLowerCase().indexOf(itext) >= 0;
        };
      },
      //location specific methods
      "nth-child": function(next, rule) {
        var func = getNCheck(rule);
        if (func === falseFunc) return func;
        if (func === trueFunc) return getChildFunc(next);
        return function nthChild(elem) {
          var siblings = getSiblings(elem);
          for (var i = 0, pos = 0; i < siblings.length; i++) {
            if (isTag(siblings[i])) {
              if (siblings[i] === elem) break;
              else pos++;
            }
          }
          return func(pos) && next(elem);
        };
      },
      "nth-last-child": function(next, rule) {
        var func = getNCheck(rule);
        if (func === falseFunc) return func;
        if (func === trueFunc) return getChildFunc(next);
        return function nthLastChild(elem) {
          var siblings = getSiblings(elem);
          for (var pos = 0, i = siblings.length - 1; i >= 0; i--) {
            if (isTag(siblings[i])) {
              if (siblings[i] === elem) break;
              else pos++;
            }
          }
          return func(pos) && next(elem);
        };
      },
      "nth-of-type": function(next, rule) {
        var func = getNCheck(rule);
        if (func === falseFunc) return func;
        if (func === trueFunc) return getChildFunc(next);
        return function nthOfType(elem) {
          var siblings = getSiblings(elem);
          for (var pos = 0, i = 0; i < siblings.length; i++) {
            if (isTag(siblings[i])) {
              if (siblings[i] === elem) break;
              if (getName(siblings[i]) === getName(elem)) pos++;
            }
          }
          return func(pos) && next(elem);
        };
      },
      "nth-last-of-type": function(next, rule) {
        var func = getNCheck(rule);
        if (func === falseFunc) return func;
        if (func === trueFunc) return getChildFunc(next);
        return function nthLastOfType(elem) {
          var siblings = getSiblings(elem);
          for (var pos = 0, i = siblings.length - 1; i >= 0; i--) {
            if (isTag(siblings[i])) {
              if (siblings[i] === elem) break;
              if (getName(siblings[i]) === getName(elem)) pos++;
            }
          }
          return func(pos) && next(elem);
        };
      },
      //TODO determine the actual root element
      root: function(next) {
        return function(elem) {
          return !getParent(elem) && next(elem);
        };
      },
      scope: function(next, rule, options, context) {
        if (!context || context.length === 0) {
          return filters.root(next);
        }
        if (context.length === 1) {
          return function(elem) {
            return context[0] === elem && next(elem);
          };
        }
        return function(elem) {
          return context.indexOf(elem) >= 0 && next(elem);
        };
      },
      //jQuery extensions (others follow as pseudos)
      checkbox: getAttribFunc("type", "checkbox"),
      file: getAttribFunc("type", "file"),
      password: getAttribFunc("type", "password"),
      radio: getAttribFunc("type", "radio"),
      reset: getAttribFunc("type", "reset"),
      image: getAttribFunc("type", "image"),
      submit: getAttribFunc("type", "submit")
    };
    var pseudos = {
      empty: function(elem) {
        return !getChildren(elem).some(function(elem2) {
          return isTag(elem2) || elem2.type === "text";
        });
      },
      "first-child": function(elem) {
        return getFirstElement(getSiblings(elem)) === elem;
      },
      "last-child": function(elem) {
        var siblings = getSiblings(elem);
        for (var i = siblings.length - 1; i >= 0; i--) {
          if (siblings[i] === elem) return true;
          if (isTag(siblings[i])) break;
        }
        return false;
      },
      "first-of-type": function(elem) {
        var siblings = getSiblings(elem);
        for (var i = 0; i < siblings.length; i++) {
          if (isTag(siblings[i])) {
            if (siblings[i] === elem) return true;
            if (getName(siblings[i]) === getName(elem)) break;
          }
        }
        return false;
      },
      "last-of-type": function(elem) {
        var siblings = getSiblings(elem);
        for (var i = siblings.length - 1; i >= 0; i--) {
          if (isTag(siblings[i])) {
            if (siblings[i] === elem) return true;
            if (getName(siblings[i]) === getName(elem)) break;
          }
        }
        return false;
      },
      "only-of-type": function(elem) {
        var siblings = getSiblings(elem);
        for (var i = 0, j = siblings.length; i < j; i++) {
          if (isTag(siblings[i])) {
            if (siblings[i] === elem) continue;
            if (getName(siblings[i]) === getName(elem)) return false;
          }
        }
        return true;
      },
      "only-child": function(elem) {
        var siblings = getSiblings(elem);
        for (var i = 0; i < siblings.length; i++) {
          if (isTag(siblings[i]) && siblings[i] !== elem) return false;
        }
        return true;
      },
      //:matches(a, area, link)[href]
      link: function(elem) {
        return hasAttrib(elem, "href");
      },
      visited: falseFunc,
      //seems to be a valid implementation
      //TODO: :any-link once the name is finalized (as an alias of :link)
      //forms
      //to consider: :target
      //:matches([selected], select:not([multiple]):not(> option[selected]) > option:first-of-type)
      selected: function(elem) {
        if (hasAttrib(elem, "selected")) return true;
        else if (getName(elem) !== "option") return false;
        var parent = getParent(elem);
        if (!parent || getName(parent) !== "select" || hasAttrib(parent, "multiple")) return false;
        var siblings = getChildren(parent), sawElem = false;
        for (var i = 0; i < siblings.length; i++) {
          if (isTag(siblings[i])) {
            if (siblings[i] === elem) {
              sawElem = true;
            } else if (!sawElem) {
              return false;
            } else if (hasAttrib(siblings[i], "selected")) {
              return false;
            }
          }
        }
        return sawElem;
      },
      //https://html.spec.whatwg.org/multipage/scripting.html#disabled-elements
      //:matches(
      //  :matches(button, input, select, textarea, menuitem, optgroup, option)[disabled],
      //  optgroup[disabled] > option),
      // fieldset[disabled] * //TODO not child of first <legend>
      //)
      disabled: function(elem) {
        return hasAttrib(elem, "disabled");
      },
      enabled: function(elem) {
        return !hasAttrib(elem, "disabled");
      },
      //:matches(:matches(:radio, :checkbox)[checked], :selected) (TODO menuitem)
      checked: function(elem) {
        return hasAttrib(elem, "checked") || pseudos.selected(elem);
      },
      //:matches(input, select, textarea)[required]
      required: function(elem) {
        return hasAttrib(elem, "required");
      },
      //:matches(input, select, textarea):not([required])
      optional: function(elem) {
        return !hasAttrib(elem, "required");
      },
      //jQuery extensions
      //:not(:empty)
      parent: function(elem) {
        return !pseudos.empty(elem);
      },
      //:matches(h1, h2, h3, h4, h5, h6)
      header: function(elem) {
        var name = getName(elem);
        return name === "h1" || name === "h2" || name === "h3" || name === "h4" || name === "h5" || name === "h6";
      },
      //:matches(button, input[type=button])
      button: function(elem) {
        var name = getName(elem);
        return name === "button" || name === "input" && getAttribute(elem, "type") === "button";
      },
      //:matches(input, textarea, select, button)
      input: function(elem) {
        var name = getName(elem);
        return name === "input" || name === "textarea" || name === "select" || name === "button";
      },
      //input:matches(:not([type!='']), [type='text' i])
      text: function(elem) {
        var attr;
        return getName(elem) === "input" && (!(attr = getAttribute(elem, "type")) || attr.toLowerCase() === "text");
      }
    };
    function verifyArgs(func, name, subselect) {
      if (subselect === null) {
        if (func.length > 1 && name !== "scope") {
          throw new SyntaxError("pseudo-selector :" + name + " requires an argument");
        }
      } else {
        if (func.length === 1) {
          throw new SyntaxError("pseudo-selector :" + name + " doesn't have any arguments");
        }
      }
    }
    var re_CSS3 = /^(?:(?:nth|last|first|only)-(?:child|of-type)|root|empty|(?:en|dis)abled|checked|not)$/;
    module2.exports = {
      compile: function(next, data, options, context) {
        var name = data.name, subselect = data.data;
        if (options && options.strict && !re_CSS3.test(name)) {
          throw SyntaxError(":" + name + " isn't part of CSS3");
        }
        if (typeof filters[name] === "function") {
          verifyArgs(filters[name], name, subselect);
          return filters[name](next, subselect, options, context);
        } else if (typeof pseudos[name] === "function") {
          var func = pseudos[name];
          verifyArgs(func, name, subselect);
          if (next === trueFunc) return func;
          return function pseudoArgs(elem) {
            return func(elem, subselect) && next(elem);
          };
        } else {
          throw new SyntaxError("unmatched pseudo-class :" + name);
        }
      },
      filters,
      pseudos
    };
  }
});

// node_modules/css-what/index.js
var require_css_what = __commonJS({
  "node_modules/css-what/index.js"(exports2, module2) {
    "use strict";
    init_process_shim();
    module2.exports = parse;
    var re_name = /^(?:\\.|[\w\-\u00b0-\uFFFF])+/;
    var re_escape = /\\([\da-f]{1,6}\s?|(\s)|.)/ig;
    var re_attr = /^\s*((?:\\.|[\w\u00b0-\uFFFF\-])+)\s*(?:(\S?)=\s*(?:(['"])([^]*?)\3|(#?(?:\\.|[\w\u00b0-\uFFFF\-])*)|)|)\s*(i)?\]/;
    var actionTypes = {
      __proto__: null,
      "undefined": "exists",
      "": "equals",
      "~": "element",
      "^": "start",
      "$": "end",
      "*": "any",
      "!": "not",
      "|": "hyphen"
    };
    var simpleSelectors = {
      __proto__: null,
      ">": "child",
      "<": "parent",
      "~": "sibling",
      "+": "adjacent"
    };
    var attribSelectors = {
      __proto__: null,
      "#": ["id", "equals"],
      ".": ["class", "element"]
    };
    var unpackPseudos = {
      __proto__: null,
      "has": true,
      "not": true,
      "matches": true
    };
    var stripQuotesFromPseudos = {
      __proto__: null,
      "contains": true,
      "icontains": true
    };
    var quotes = {
      __proto__: null,
      '"': true,
      "'": true
    };
    function funescape(_, escaped, escapedWhitespace) {
      var high = "0x" + escaped - 65536;
      return high !== high || escapedWhitespace ? escaped : (
        // BMP codepoint
        high < 0 ? String.fromCharCode(high + 65536) : (
          // Supplemental Plane codepoint (surrogate pair)
          String.fromCharCode(high >> 10 | 55296, high & 1023 | 56320)
        )
      );
    }
    function unescapeCSS(str) {
      return str.replace(re_escape, funescape);
    }
    function isWhitespace(c) {
      return c === " " || c === "\n" || c === "	" || c === "\f" || c === "\r";
    }
    function parse(selector, options) {
      var subselects = [];
      selector = parseSelector(subselects, selector + "", options);
      if (selector !== "") {
        throw new SyntaxError("Unmatched selector: " + selector);
      }
      return subselects;
    }
    function parseSelector(subselects, selector, options) {
      var tokens = [], sawWS = false, data, firstChar, name, quot;
      function getName() {
        var sub = selector.match(re_name)[0];
        selector = selector.substr(sub.length);
        return unescapeCSS(sub);
      }
      function stripWhitespace(start) {
        while (isWhitespace(selector.charAt(start))) start++;
        selector = selector.substr(start);
      }
      function isEscaped(pos2) {
        var slashCount = 0;
        while (selector.charAt(--pos2) === "\\") slashCount++;
        return (slashCount & 1) === 1;
      }
      stripWhitespace(0);
      while (selector !== "") {
        firstChar = selector.charAt(0);
        if (isWhitespace(firstChar)) {
          sawWS = true;
          stripWhitespace(1);
        } else if (firstChar in simpleSelectors) {
          tokens.push({ type: simpleSelectors[firstChar] });
          sawWS = false;
          stripWhitespace(1);
        } else if (firstChar === ",") {
          if (tokens.length === 0) {
            throw new SyntaxError("empty sub-selector");
          }
          subselects.push(tokens);
          tokens = [];
          sawWS = false;
          stripWhitespace(1);
        } else {
          if (sawWS) {
            if (tokens.length > 0) {
              tokens.push({ type: "descendant" });
            }
            sawWS = false;
          }
          if (firstChar === "*") {
            selector = selector.substr(1);
            tokens.push({ type: "universal" });
          } else if (firstChar in attribSelectors) {
            selector = selector.substr(1);
            tokens.push({
              type: "attribute",
              name: attribSelectors[firstChar][0],
              action: attribSelectors[firstChar][1],
              value: getName(),
              ignoreCase: false
            });
          } else if (firstChar === "[") {
            selector = selector.substr(1);
            data = selector.match(re_attr);
            if (!data) {
              throw new SyntaxError("Malformed attribute selector: " + selector);
            }
            selector = selector.substr(data[0].length);
            name = unescapeCSS(data[1]);
            if (!options || ("lowerCaseAttributeNames" in options ? options.lowerCaseAttributeNames : !options.xmlMode)) {
              name = name.toLowerCase();
            }
            tokens.push({
              type: "attribute",
              name,
              action: actionTypes[data[2]],
              value: unescapeCSS(data[4] || data[5] || ""),
              ignoreCase: !!data[6]
            });
          } else if (firstChar === ":") {
            if (selector.charAt(1) === ":") {
              selector = selector.substr(2);
              tokens.push({ type: "pseudo-element", name: getName().toLowerCase() });
              continue;
            }
            selector = selector.substr(1);
            name = getName().toLowerCase();
            data = null;
            if (selector.charAt(0) === "(") {
              if (name in unpackPseudos) {
                quot = selector.charAt(1);
                var quoted = quot in quotes;
                selector = selector.substr(quoted + 1);
                data = [];
                selector = parseSelector(data, selector, options);
                if (quoted) {
                  if (selector.charAt(0) !== quot) {
                    throw new SyntaxError("unmatched quotes in :" + name);
                  } else {
                    selector = selector.substr(1);
                  }
                }
                if (selector.charAt(0) !== ")") {
                  throw new SyntaxError("missing closing parenthesis in :" + name + " " + selector);
                }
                selector = selector.substr(1);
              } else {
                var pos = 1, counter = 1;
                for (; counter > 0 && pos < selector.length; pos++) {
                  if (selector.charAt(pos) === "(" && !isEscaped(pos)) counter++;
                  else if (selector.charAt(pos) === ")" && !isEscaped(pos)) counter--;
                }
                if (counter) {
                  throw new SyntaxError("parenthesis not matched");
                }
                data = selector.substr(1, pos - 2);
                selector = selector.substr(pos);
                if (name in stripQuotesFromPseudos) {
                  quot = data.charAt(0);
                  if (quot === data.slice(-1) && quot in quotes) {
                    data = data.slice(1, -1);
                  }
                  data = unescapeCSS(data);
                }
              }
            }
            tokens.push({ type: "pseudo", name, data });
          } else if (re_name.test(selector)) {
            name = getName();
            if (!options || ("lowerCaseTags" in options ? options.lowerCaseTags : !options.xmlMode)) {
              name = name.toLowerCase();
            }
            tokens.push({ type: "tag", name });
          } else {
            if (tokens.length && tokens[tokens.length - 1].type === "descendant") {
              tokens.pop();
            }
            addToken(subselects, tokens);
            return selector;
          }
        }
      }
      addToken(subselects, tokens);
      return selector;
    }
    function addToken(subselects, tokens) {
      if (subselects.length > 0 && tokens.length === 0) {
        throw new SyntaxError("empty sub-selector");
      }
      subselects.push(tokens);
    }
  }
});

// node_modules/css-select/lib/general.js
var require_general = __commonJS({
  "node_modules/css-select/lib/general.js"(exports2, module2) {
    init_process_shim();
    var DomUtils = require_domutils();
    var isTag = DomUtils.isTag;
    var getParent = DomUtils.getParent;
    var getChildren = DomUtils.getChildren;
    var getSiblings = DomUtils.getSiblings;
    var getName = DomUtils.getName;
    module2.exports = {
      __proto__: null,
      attribute: require_attributes().compile,
      pseudo: require_pseudos().compile,
      //tags
      tag: function(next, data) {
        var name = data.name;
        return function tag(elem) {
          return getName(elem) === name && next(elem);
        };
      },
      //traversal
      descendant: function(next, rule, options, context, acceptSelf) {
        return function descendant(elem) {
          if (acceptSelf && next(elem)) return true;
          var found = false;
          while (!found && (elem = getParent(elem))) {
            found = next(elem);
          }
          return found;
        };
      },
      parent: function(next, data, options) {
        if (options && options.strict) throw SyntaxError("Parent selector isn't part of CSS3");
        return function parent(elem) {
          return getChildren(elem).some(test);
        };
        function test(elem) {
          return isTag(elem) && next(elem);
        }
      },
      child: function(next) {
        return function child(elem) {
          var parent = getParent(elem);
          return !!parent && next(parent);
        };
      },
      sibling: function(next) {
        return function sibling(elem) {
          var siblings = getSiblings(elem);
          for (var i = 0; i < siblings.length; i++) {
            if (isTag(siblings[i])) {
              if (siblings[i] === elem) break;
              if (next(siblings[i])) return true;
            }
          }
          return false;
        };
      },
      adjacent: function(next) {
        return function adjacent(elem) {
          var siblings = getSiblings(elem), lastElement;
          for (var i = 0; i < siblings.length; i++) {
            if (isTag(siblings[i])) {
              if (siblings[i] === elem) break;
              lastElement = siblings[i];
            }
          }
          return !!lastElement && next(lastElement);
        };
      },
      universal: function(next) {
        return next;
      }
    };
  }
});

// node_modules/css-select/lib/procedure.json
var require_procedure = __commonJS({
  "node_modules/css-select/lib/procedure.json"(exports2, module2) {
    module2.exports = {
      universal: 50,
      tag: 30,
      attribute: 1,
      pseudo: 0,
      descendant: -1,
      child: -1,
      parent: -1,
      sibling: -1,
      adjacent: -1
    };
  }
});

// node_modules/css-select/lib/sort.js
var require_sort = __commonJS({
  "node_modules/css-select/lib/sort.js"(exports2, module2) {
    init_process_shim();
    module2.exports = sortByProcedure;
    var procedure = require_procedure();
    var attributes = {
      __proto__: null,
      exists: 10,
      equals: 8,
      not: 7,
      start: 6,
      end: 6,
      any: 5,
      hyphen: 4,
      element: 4
    };
    function sortByProcedure(arr) {
      var procs = arr.map(getProcedure);
      for (var i = 1; i < arr.length; i++) {
        var procNew = procs[i];
        if (procNew < 0) continue;
        for (var j = i - 1; j >= 0 && procNew < procs[j]; j--) {
          var token = arr[j + 1];
          arr[j + 1] = arr[j];
          arr[j] = token;
          procs[j + 1] = procs[j];
          procs[j] = procNew;
        }
      }
    }
    function getProcedure(token) {
      var proc = procedure[token.type];
      if (proc === procedure.attribute) {
        proc = attributes[token.action];
        if (proc === attributes.equals && token.name === "id") {
          proc = 9;
        }
        if (token.ignoreCase) {
          proc >>= 1;
        }
      } else if (proc === procedure.pseudo) {
        if (!token.data) {
          proc = 3;
        } else if (token.name === "has" || token.name === "contains") {
          proc = 0;
        } else if (token.name === "matches" || token.name === "not") {
          proc = 0;
          for (var i = 0; i < token.data.length; i++) {
            if (token.data[i].length !== 1) continue;
            var cur = getProcedure(token.data[i][0]);
            if (cur === 0) {
              proc = 0;
              break;
            }
            if (cur > proc) proc = cur;
          }
          if (token.data.length > 1 && proc > 0) proc -= 1;
        } else {
          proc = 1;
        }
      }
      return proc;
    }
  }
});

// node_modules/css-select/lib/compile.js
var require_compile2 = __commonJS({
  "node_modules/css-select/lib/compile.js"(exports2, module2) {
    init_process_shim();
    module2.exports = compile;
    module2.exports.compileUnsafe = compileUnsafe;
    module2.exports.compileToken = compileToken;
    var parse = require_css_what();
    var DomUtils = require_domutils();
    var isTag = DomUtils.isTag;
    var Rules = require_general();
    var sortRules = require_sort();
    var BaseFuncs = require_boolbase();
    var trueFunc = BaseFuncs.trueFunc;
    var falseFunc = BaseFuncs.falseFunc;
    var procedure = require_procedure();
    function compile(selector, options, context) {
      var next = compileUnsafe(selector, options, context);
      return wrap(next);
    }
    function wrap(next) {
      return function base(elem) {
        return isTag(elem) && next(elem);
      };
    }
    function compileUnsafe(selector, options, context) {
      var token = parse(selector, options);
      return compileToken(token, options, context);
    }
    function includesScopePseudo(t) {
      return t.type === "pseudo" && (t.name === "scope" || Array.isArray(t.data) && t.data.some(function(data) {
        return data.some(includesScopePseudo);
      }));
    }
    var DESCENDANT_TOKEN = { type: "descendant" };
    var SCOPE_TOKEN = { type: "pseudo", name: "scope" };
    var PLACEHOLDER_ELEMENT = {};
    var getParent = DomUtils.getParent;
    function absolutize(token, context) {
      var hasContext = !!context && !!context.length && context.every(function(e) {
        return e === PLACEHOLDER_ELEMENT || !!getParent(e);
      });
      token.forEach(function(t) {
        if (t.length > 0 && isTraversal(t[0]) && t[0].type !== "descendant") {
        } else if (hasContext && !includesScopePseudo(t)) {
          t.unshift(DESCENDANT_TOKEN);
        } else {
          return;
        }
        t.unshift(SCOPE_TOKEN);
      });
    }
    function compileToken(token, options, context) {
      token = token.filter(function(t) {
        return t.length > 0;
      });
      token.forEach(sortRules);
      var isArrayContext = Array.isArray(context);
      context = options && options.context || context;
      if (context && !isArrayContext) context = [context];
      absolutize(token, context);
      return token.map(function(rules) {
        return compileRules(rules, options, context, isArrayContext);
      }).reduce(reduceRules, falseFunc);
    }
    function isTraversal(t) {
      return procedure[t.type] < 0;
    }
    function compileRules(rules, options, context, isArrayContext) {
      var acceptSelf = isArrayContext && rules[0].name === "scope" && rules[1].type === "descendant";
      return rules.reduce(function(func, rule, index) {
        if (func === falseFunc) return func;
        return Rules[rule.type](func, rule, options, context, acceptSelf && index === 1);
      }, options && options.rootFunc || trueFunc);
    }
    function reduceRules(a, b) {
      if (b === falseFunc || a === trueFunc) {
        return a;
      }
      if (a === falseFunc || b === trueFunc) {
        return b;
      }
      return function combine(elem) {
        return a(elem) || b(elem);
      };
    }
    var Pseudos = require_pseudos();
    var filters = Pseudos.filters;
    var existsOne = DomUtils.existsOne;
    var isTag = DomUtils.isTag;
    var getChildren = DomUtils.getChildren;
    function containsTraversal(t) {
      return t.some(isTraversal);
    }
    filters.not = function(next, token, options, context) {
      var opts = {
        xmlMode: !!(options && options.xmlMode),
        strict: !!(options && options.strict)
      };
      if (opts.strict) {
        if (token.length > 1 || token.some(containsTraversal)) {
          throw new SyntaxError("complex selectors in :not aren't allowed in strict mode");
        }
      }
      var func = compileToken(token, opts, context);
      if (func === falseFunc) return next;
      if (func === trueFunc) return falseFunc;
      return function(elem) {
        return !func(elem) && next(elem);
      };
    };
    filters.has = function(next, token, options) {
      var opts = {
        xmlMode: !!(options && options.xmlMode),
        strict: !!(options && options.strict)
      };
      var context = token.some(containsTraversal) ? [PLACEHOLDER_ELEMENT] : null;
      var func = compileToken(token, opts, context);
      if (func === falseFunc) return falseFunc;
      if (func === trueFunc) return function(elem) {
        return getChildren(elem).some(isTag) && next(elem);
      };
      func = wrap(func);
      if (context) {
        return function has(elem) {
          return next(elem) && (context[0] = elem, existsOne(func, getChildren(elem)));
        };
      }
      return function has(elem) {
        return next(elem) && existsOne(func, getChildren(elem));
      };
    };
    filters.matches = function(next, token, options, context) {
      var opts = {
        xmlMode: !!(options && options.xmlMode),
        strict: !!(options && options.strict),
        rootFunc: next
      };
      return compileToken(token, opts, context);
    };
  }
});

// node_modules/css-select/index.js
var require_css_select = __commonJS({
  "node_modules/css-select/index.js"(exports2, module2) {
    "use strict";
    init_process_shim();
    module2.exports = CSSselect;
    var Pseudos = require_pseudos();
    var DomUtils = require_domutils();
    var findOne = DomUtils.findOne;
    var findAll = DomUtils.findAll;
    var getChildren = DomUtils.getChildren;
    var removeSubsets = DomUtils.removeSubsets;
    var falseFunc = require_boolbase().falseFunc;
    var compile = require_compile2();
    var compileUnsafe = compile.compileUnsafe;
    var compileToken = compile.compileToken;
    function getSelectorFunc(searchFunc) {
      return function select(query, elems, options) {
        if (typeof query !== "function") query = compileUnsafe(query, options, elems);
        if (!Array.isArray(elems)) elems = getChildren(elems);
        else elems = removeSubsets(elems);
        return searchFunc(query, elems);
      };
    }
    var selectAll = getSelectorFunc(function selectAll2(query, elems) {
      return query === falseFunc || !elems || elems.length === 0 ? [] : findAll(query, elems);
    });
    var selectOne = getSelectorFunc(function selectOne2(query, elems) {
      return query === falseFunc || !elems || elems.length === 0 ? null : findOne(query, elems);
    });
    function is(elem, query, options) {
      return (typeof query === "function" ? query : compile(query, options))(elem);
    }
    function CSSselect(query, elems, options) {
      return selectAll(query, elems, options);
    }
    CSSselect.compile = compile;
    CSSselect.filters = Pseudos.filters;
    CSSselect.pseudos = Pseudos.pseudos;
    CSSselect.selectAll = selectAll;
    CSSselect.selectOne = selectOne;
    CSSselect.is = is;
    CSSselect.parse = compile;
    CSSselect.iterate = selectAll;
    CSSselect._compileUnsafe = compileUnsafe;
    CSSselect._compileToken = compileToken;
  }
});

// node_modules/lodash.merge/index.js
var require_lodash5 = __commonJS({
  "node_modules/lodash.merge/index.js"(exports2, module2) {
    init_process_shim();
    var LARGE_ARRAY_SIZE = 200;
    var HASH_UNDEFINED = "__lodash_hash_undefined__";
    var HOT_COUNT = 800;
    var HOT_SPAN = 16;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var arrayTag = "[object Array]";
    var asyncTag = "[object AsyncFunction]";
    var boolTag = "[object Boolean]";
    var dateTag = "[object Date]";
    var errorTag = "[object Error]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var mapTag = "[object Map]";
    var numberTag = "[object Number]";
    var nullTag = "[object Null]";
    var objectTag = "[object Object]";
    var proxyTag = "[object Proxy]";
    var regexpTag = "[object RegExp]";
    var setTag = "[object Set]";
    var stringTag = "[object String]";
    var undefinedTag = "[object Undefined]";
    var weakMapTag = "[object WeakMap]";
    var arrayBufferTag = "[object ArrayBuffer]";
    var dataViewTag = "[object DataView]";
    var float32Tag = "[object Float32Array]";
    var float64Tag = "[object Float64Array]";
    var int8Tag = "[object Int8Array]";
    var int16Tag = "[object Int16Array]";
    var int32Tag = "[object Int32Array]";
    var uint8Tag = "[object Uint8Array]";
    var uint8ClampedTag = "[object Uint8ClampedArray]";
    var uint16Tag = "[object Uint16Array]";
    var uint32Tag = "[object Uint32Array]";
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var typedArrayTags = {};
    typedArrayTags[float32Tag] = typedArrayTags[float64Tag] = typedArrayTags[int8Tag] = typedArrayTags[int16Tag] = typedArrayTags[int32Tag] = typedArrayTags[uint8Tag] = typedArrayTags[uint8ClampedTag] = typedArrayTags[uint16Tag] = typedArrayTags[uint32Tag] = true;
    typedArrayTags[argsTag] = typedArrayTags[arrayTag] = typedArrayTags[arrayBufferTag] = typedArrayTags[boolTag] = typedArrayTags[dataViewTag] = typedArrayTags[dateTag] = typedArrayTags[errorTag] = typedArrayTags[funcTag] = typedArrayTags[mapTag] = typedArrayTags[numberTag] = typedArrayTags[objectTag] = typedArrayTags[regexpTag] = typedArrayTags[setTag] = typedArrayTags[stringTag] = typedArrayTags[weakMapTag] = false;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    var freeExports = typeof exports2 == "object" && exports2 && !exports2.nodeType && exports2;
    var freeModule = freeExports && typeof module2 == "object" && module2 && !module2.nodeType && module2;
    var moduleExports = freeModule && freeModule.exports === freeExports;
    var freeProcess = moduleExports && freeGlobal.process;
    var nodeUtil = (function() {
      try {
        var types = freeModule && freeModule.require && freeModule.require("util").types;
        if (types) {
          return types;
        }
        return freeProcess && freeProcess.binding && freeProcess.binding("util");
      } catch (e) {
      }
    })();
    var nodeIsTypedArray = nodeUtil && nodeUtil.isTypedArray;
    function apply(func, thisArg, args) {
      switch (args.length) {
        case 0:
          return func.call(thisArg);
        case 1:
          return func.call(thisArg, args[0]);
        case 2:
          return func.call(thisArg, args[0], args[1]);
        case 3:
          return func.call(thisArg, args[0], args[1], args[2]);
      }
      return func.apply(thisArg, args);
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function baseUnary(func) {
      return function(value) {
        return func(value);
      };
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    var arrayProto = Array.prototype;
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var nativeObjectToString = objectProto.toString;
    var objectCtorString = funcToString.call(Object);
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var Buffer2 = moduleExports ? root.Buffer : void 0;
    var Symbol2 = root.Symbol;
    var Uint8Array2 = root.Uint8Array;
    var allocUnsafe = Buffer2 ? Buffer2.allocUnsafe : void 0;
    var getPrototype = overArg(Object.getPrototypeOf, Object);
    var objectCreate = Object.create;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var splice = arrayProto.splice;
    var symToStringTag = Symbol2 ? Symbol2.toStringTag : void 0;
    var defineProperty = (function() {
      try {
        var func = getNative(Object, "defineProperty");
        func({}, "", {});
        return func;
      } catch (e) {
      }
    })();
    var nativeIsBuffer = Buffer2 ? Buffer2.isBuffer : void 0;
    var nativeMax = Math.max;
    var nativeNow = Date.now;
    var Map = getNative(root, "Map");
    var nativeCreate = getNative(Object, "create");
    var baseCreate = /* @__PURE__ */ (function() {
      function object() {
      }
      return function(proto) {
        if (!isObject(proto)) {
          return {};
        }
        if (objectCreate) {
          return objectCreate(proto);
        }
        object.prototype = proto;
        var result = new object();
        object.prototype = void 0;
        return result;
      };
    })();
    function Hash(entries) {
      var index = -1, length = entries == null ? 0 : entries.length;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function hashClear() {
      this.__data__ = nativeCreate ? nativeCreate(null) : {};
      this.size = 0;
    }
    function hashDelete(key) {
      var result = this.has(key) && delete this.__data__[key];
      this.size -= result ? 1 : 0;
      return result;
    }
    function hashGet(key) {
      var data = this.__data__;
      if (nativeCreate) {
        var result = data[key];
        return result === HASH_UNDEFINED ? void 0 : result;
      }
      return hasOwnProperty2.call(data, key) ? data[key] : void 0;
    }
    function hashHas(key) {
      var data = this.__data__;
      return nativeCreate ? data[key] !== void 0 : hasOwnProperty2.call(data, key);
    }
    function hashSet(key, value) {
      var data = this.__data__;
      this.size += this.has(key) ? 0 : 1;
      data[key] = nativeCreate && value === void 0 ? HASH_UNDEFINED : value;
      return this;
    }
    Hash.prototype.clear = hashClear;
    Hash.prototype["delete"] = hashDelete;
    Hash.prototype.get = hashGet;
    Hash.prototype.has = hashHas;
    Hash.prototype.set = hashSet;
    function ListCache(entries) {
      var index = -1, length = entries == null ? 0 : entries.length;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function listCacheClear() {
      this.__data__ = [];
      this.size = 0;
    }
    function listCacheDelete(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        return false;
      }
      var lastIndex = data.length - 1;
      if (index == lastIndex) {
        data.pop();
      } else {
        splice.call(data, index, 1);
      }
      --this.size;
      return true;
    }
    function listCacheGet(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      return index < 0 ? void 0 : data[index][1];
    }
    function listCacheHas(key) {
      return assocIndexOf(this.__data__, key) > -1;
    }
    function listCacheSet(key, value) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        ++this.size;
        data.push([key, value]);
      } else {
        data[index][1] = value;
      }
      return this;
    }
    ListCache.prototype.clear = listCacheClear;
    ListCache.prototype["delete"] = listCacheDelete;
    ListCache.prototype.get = listCacheGet;
    ListCache.prototype.has = listCacheHas;
    ListCache.prototype.set = listCacheSet;
    function MapCache(entries) {
      var index = -1, length = entries == null ? 0 : entries.length;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function mapCacheClear() {
      this.size = 0;
      this.__data__ = {
        "hash": new Hash(),
        "map": new (Map || ListCache)(),
        "string": new Hash()
      };
    }
    function mapCacheDelete(key) {
      var result = getMapData(this, key)["delete"](key);
      this.size -= result ? 1 : 0;
      return result;
    }
    function mapCacheGet(key) {
      return getMapData(this, key).get(key);
    }
    function mapCacheHas(key) {
      return getMapData(this, key).has(key);
    }
    function mapCacheSet(key, value) {
      var data = getMapData(this, key), size = data.size;
      data.set(key, value);
      this.size += data.size == size ? 0 : 1;
      return this;
    }
    MapCache.prototype.clear = mapCacheClear;
    MapCache.prototype["delete"] = mapCacheDelete;
    MapCache.prototype.get = mapCacheGet;
    MapCache.prototype.has = mapCacheHas;
    MapCache.prototype.set = mapCacheSet;
    function Stack(entries) {
      var data = this.__data__ = new ListCache(entries);
      this.size = data.size;
    }
    function stackClear() {
      this.__data__ = new ListCache();
      this.size = 0;
    }
    function stackDelete(key) {
      var data = this.__data__, result = data["delete"](key);
      this.size = data.size;
      return result;
    }
    function stackGet(key) {
      return this.__data__.get(key);
    }
    function stackHas(key) {
      return this.__data__.has(key);
    }
    function stackSet(key, value) {
      var data = this.__data__;
      if (data instanceof ListCache) {
        var pairs = data.__data__;
        if (!Map || pairs.length < LARGE_ARRAY_SIZE - 1) {
          pairs.push([key, value]);
          this.size = ++data.size;
          return this;
        }
        data = this.__data__ = new MapCache(pairs);
      }
      data.set(key, value);
      this.size = data.size;
      return this;
    }
    Stack.prototype.clear = stackClear;
    Stack.prototype["delete"] = stackDelete;
    Stack.prototype.get = stackGet;
    Stack.prototype.has = stackHas;
    Stack.prototype.set = stackSet;
    function arrayLikeKeys(value, inherited) {
      var isArr = isArray(value), isArg = !isArr && isArguments(value), isBuff = !isArr && !isArg && isBuffer(value), isType = !isArr && !isArg && !isBuff && isTypedArray(value), skipIndexes = isArr || isArg || isBuff || isType, result = skipIndexes ? baseTimes(value.length, String) : [], length = result.length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && // Safari 9 has enumerable `arguments.length` in strict mode.
        (key == "length" || // Node.js 0.10 has enumerable non-index properties on buffers.
        isBuff && (key == "offset" || key == "parent") || // PhantomJS 2 has enumerable non-index properties on typed arrays.
        isType && (key == "buffer" || key == "byteLength" || key == "byteOffset") || // Skip index properties.
        isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assignMergeValue(object, key, value) {
      if (value !== void 0 && !eq(object[key], value) || value === void 0 && !(key in object)) {
        baseAssignValue(object, key, value);
      }
    }
    function assignValue(object, key, value) {
      var objValue = object[key];
      if (!(hasOwnProperty2.call(object, key) && eq(objValue, value)) || value === void 0 && !(key in object)) {
        baseAssignValue(object, key, value);
      }
    }
    function assocIndexOf(array, key) {
      var length = array.length;
      while (length--) {
        if (eq(array[length][0], key)) {
          return length;
        }
      }
      return -1;
    }
    function baseAssignValue(object, key, value) {
      if (key == "__proto__" && defineProperty) {
        defineProperty(object, key, {
          "configurable": true,
          "enumerable": true,
          "value": value,
          "writable": true
        });
      } else {
        object[key] = value;
      }
    }
    var baseFor = createBaseFor();
    function baseGetTag(value) {
      if (value == null) {
        return value === void 0 ? undefinedTag : nullTag;
      }
      return symToStringTag && symToStringTag in Object(value) ? getRawTag(value) : objectToString(value);
    }
    function baseIsArguments(value) {
      return isObjectLike(value) && baseGetTag(value) == argsTag;
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseIsTypedArray(value) {
      return isObjectLike(value) && isLength(value.length) && !!typedArrayTags[baseGetTag(value)];
    }
    function baseKeysIn(object) {
      if (!isObject(object)) {
        return nativeKeysIn(object);
      }
      var isProto = isPrototype(object), result = [];
      for (var key in object) {
        if (!(key == "constructor" && (isProto || !hasOwnProperty2.call(object, key)))) {
          result.push(key);
        }
      }
      return result;
    }
    function baseMerge(object, source, srcIndex, customizer, stack) {
      if (object === source) {
        return;
      }
      baseFor(source, function(srcValue, key) {
        stack || (stack = new Stack());
        if (isObject(srcValue)) {
          baseMergeDeep(object, source, key, srcIndex, baseMerge, customizer, stack);
        } else {
          var newValue = customizer ? customizer(safeGet(object, key), srcValue, key + "", object, source, stack) : void 0;
          if (newValue === void 0) {
            newValue = srcValue;
          }
          assignMergeValue(object, key, newValue);
        }
      }, keysIn);
    }
    function baseMergeDeep(object, source, key, srcIndex, mergeFunc, customizer, stack) {
      var objValue = safeGet(object, key), srcValue = safeGet(source, key), stacked = stack.get(srcValue);
      if (stacked) {
        assignMergeValue(object, key, stacked);
        return;
      }
      var newValue = customizer ? customizer(objValue, srcValue, key + "", object, source, stack) : void 0;
      var isCommon = newValue === void 0;
      if (isCommon) {
        var isArr = isArray(srcValue), isBuff = !isArr && isBuffer(srcValue), isTyped = !isArr && !isBuff && isTypedArray(srcValue);
        newValue = srcValue;
        if (isArr || isBuff || isTyped) {
          if (isArray(objValue)) {
            newValue = objValue;
          } else if (isArrayLikeObject(objValue)) {
            newValue = copyArray(objValue);
          } else if (isBuff) {
            isCommon = false;
            newValue = cloneBuffer(srcValue, true);
          } else if (isTyped) {
            isCommon = false;
            newValue = cloneTypedArray(srcValue, true);
          } else {
            newValue = [];
          }
        } else if (isPlainObject(srcValue) || isArguments(srcValue)) {
          newValue = objValue;
          if (isArguments(objValue)) {
            newValue = toPlainObject(objValue);
          } else if (!isObject(objValue) || isFunction(objValue)) {
            newValue = initCloneObject(srcValue);
          }
        } else {
          isCommon = false;
        }
      }
      if (isCommon) {
        stack.set(srcValue, newValue);
        mergeFunc(newValue, srcValue, srcIndex, customizer, stack);
        stack["delete"](srcValue);
      }
      assignMergeValue(object, key, newValue);
    }
    function baseRest(func, start) {
      return setToString(overRest(func, start, identity), func + "");
    }
    var baseSetToString = !defineProperty ? identity : function(func, string) {
      return defineProperty(func, "toString", {
        "configurable": true,
        "enumerable": false,
        "value": constant(string),
        "writable": true
      });
    };
    function cloneBuffer(buffer, isDeep) {
      if (isDeep) {
        return buffer.slice();
      }
      var length = buffer.length, result = allocUnsafe ? allocUnsafe(length) : new buffer.constructor(length);
      buffer.copy(result);
      return result;
    }
    function cloneArrayBuffer(arrayBuffer) {
      var result = new arrayBuffer.constructor(arrayBuffer.byteLength);
      new Uint8Array2(result).set(new Uint8Array2(arrayBuffer));
      return result;
    }
    function cloneTypedArray(typedArray, isDeep) {
      var buffer = isDeep ? cloneArrayBuffer(typedArray.buffer) : typedArray.buffer;
      return new typedArray.constructor(buffer, typedArray.byteOffset, typedArray.length);
    }
    function copyArray(source, array) {
      var index = -1, length = source.length;
      array || (array = Array(length));
      while (++index < length) {
        array[index] = source[index];
      }
      return array;
    }
    function copyObject(source, props, object, customizer) {
      var isNew = !object;
      object || (object = {});
      var index = -1, length = props.length;
      while (++index < length) {
        var key = props[index];
        var newValue = customizer ? customizer(object[key], source[key], key, object, source) : void 0;
        if (newValue === void 0) {
          newValue = source[key];
        }
        if (isNew) {
          baseAssignValue(object, key, newValue);
        } else {
          assignValue(object, key, newValue);
        }
      }
      return object;
    }
    function createAssigner(assigner) {
      return baseRest(function(object, sources) {
        var index = -1, length = sources.length, customizer = length > 1 ? sources[length - 1] : void 0, guard = length > 2 ? sources[2] : void 0;
        customizer = assigner.length > 3 && typeof customizer == "function" ? (length--, customizer) : void 0;
        if (guard && isIterateeCall(sources[0], sources[1], guard)) {
          customizer = length < 3 ? void 0 : customizer;
          length = 1;
        }
        object = Object(object);
        while (++index < length) {
          var source = sources[index];
          if (source) {
            assigner(object, source, index, customizer);
          }
        }
        return object;
      });
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function getMapData(map, key) {
      var data = map.__data__;
      return isKeyable(key) ? data[typeof key == "string" ? "string" : "hash"] : data.map;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    function getRawTag(value) {
      var isOwn = hasOwnProperty2.call(value, symToStringTag), tag = value[symToStringTag];
      try {
        value[symToStringTag] = void 0;
        var unmasked = true;
      } catch (e) {
      }
      var result = nativeObjectToString.call(value);
      if (unmasked) {
        if (isOwn) {
          value[symToStringTag] = tag;
        } else {
          delete value[symToStringTag];
        }
      }
      return result;
    }
    function initCloneObject(object) {
      return typeof object.constructor == "function" && !isPrototype(object) ? baseCreate(getPrototype(object)) : {};
    }
    function isIndex(value, length) {
      var type = typeof value;
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (type == "number" || type != "symbol" && reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isIterateeCall(value, index, object) {
      if (!isObject(object)) {
        return false;
      }
      var type = typeof index;
      if (type == "number" ? isArrayLike(object) && isIndex(index, object.length) : type == "string" && index in object) {
        return eq(object[index], value);
      }
      return false;
    }
    function isKeyable(value) {
      var type = typeof value;
      return type == "string" || type == "number" || type == "symbol" || type == "boolean" ? value !== "__proto__" : value === null;
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function nativeKeysIn(object) {
      var result = [];
      if (object != null) {
        for (var key in Object(object)) {
          result.push(key);
        }
      }
      return result;
    }
    function objectToString(value) {
      return nativeObjectToString.call(value);
    }
    function overRest(func, start, transform) {
      start = nativeMax(start === void 0 ? func.length - 1 : start, 0);
      return function() {
        var args = arguments, index = -1, length = nativeMax(args.length - start, 0), array = Array(length);
        while (++index < length) {
          array[index] = args[start + index];
        }
        index = -1;
        var otherArgs = Array(start + 1);
        while (++index < start) {
          otherArgs[index] = args[index];
        }
        otherArgs[start] = transform(array);
        return apply(func, this, otherArgs);
      };
    }
    function safeGet(object, key) {
      if (key === "constructor" && typeof object[key] === "function") {
        return;
      }
      if (key == "__proto__") {
        return;
      }
      return object[key];
    }
    var setToString = shortOut(baseSetToString);
    function shortOut(func) {
      var count = 0, lastCalled = 0;
      return function() {
        var stamp = nativeNow(), remaining = HOT_SPAN - (stamp - lastCalled);
        lastCalled = stamp;
        if (remaining > 0) {
          if (++count >= HOT_COUNT) {
            return arguments[0];
          }
        } else {
          count = 0;
        }
        return func.apply(void 0, arguments);
      };
    }
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    var isArguments = baseIsArguments(/* @__PURE__ */ (function() {
      return arguments;
    })()) ? baseIsArguments : function(value) {
      return isObjectLike(value) && hasOwnProperty2.call(value, "callee") && !propertyIsEnumerable.call(value, "callee");
    };
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    var isBuffer = nativeIsBuffer || stubFalse;
    function isFunction(value) {
      if (!isObject(value)) {
        return false;
      }
      var tag = baseGetTag(value);
      return tag == funcTag || tag == genTag || tag == asyncTag || tag == proxyTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return value != null && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return value != null && typeof value == "object";
    }
    function isPlainObject(value) {
      if (!isObjectLike(value) || baseGetTag(value) != objectTag) {
        return false;
      }
      var proto = getPrototype(value);
      if (proto === null) {
        return true;
      }
      var Ctor = hasOwnProperty2.call(proto, "constructor") && proto.constructor;
      return typeof Ctor == "function" && Ctor instanceof Ctor && funcToString.call(Ctor) == objectCtorString;
    }
    var isTypedArray = nodeIsTypedArray ? baseUnary(nodeIsTypedArray) : baseIsTypedArray;
    function toPlainObject(value) {
      return copyObject(value, keysIn(value));
    }
    function keysIn(object) {
      return isArrayLike(object) ? arrayLikeKeys(object, true) : baseKeysIn(object);
    }
    var merge = createAssigner(function(object, source, srcIndex) {
      baseMerge(object, source, srcIndex);
    });
    function constant(value) {
      return function() {
        return value;
      };
    }
    function identity(value) {
      return value;
    }
    function stubFalse() {
      return false;
    }
    module2.exports = merge;
  }
});

// node_modules/cheerio/lib/static.js
var require_static = __commonJS({
  "node_modules/cheerio/lib/static.js"(exports2) {
    init_process_shim();
    var serialize = require_dom_serializer();
    var select = require_css_select();
    var parse = require_parse();
    var _ = {
      merge: require_lodash5(),
      defaults: require_lodash4()
    };
    exports2.load = function(content, options) {
      var Cheerio = require_cheerio();
      options = _.defaults(options || {}, Cheerio.prototype.options);
      var root = parse(content, options);
      var initialize = function(selector, context, r, opts) {
        if (!(this instanceof initialize)) {
          return new initialize(selector, context, r, opts);
        }
        opts = _.defaults(opts || {}, options);
        return Cheerio.call(this, selector, context, r || root, opts);
      };
      initialize.prototype = Object.create(Cheerio.prototype);
      initialize.prototype.constructor = initialize;
      initialize.fn = initialize.prototype;
      initialize.prototype._originalRoot = root;
      _.merge(initialize, exports2);
      initialize._root = root;
      initialize._options = options;
      return initialize;
    };
    function render(that, dom, options) {
      if (!dom) {
        if (that._root && that._root.children) {
          dom = that._root.children;
        } else {
          return "";
        }
      } else if (typeof dom === "string") {
        dom = select(dom, that._root, options);
      }
      return serialize(dom, options);
    }
    exports2.html = function(dom, options) {
      var Cheerio = require_cheerio();
      if (Object.prototype.toString.call(dom) === "[object Object]" && !options && !("length" in dom) && !("type" in dom)) {
        options = dom;
        dom = void 0;
      }
      options = _.defaults(options || {}, this._options, Cheerio.prototype.options);
      return render(this, dom, options);
    };
    exports2.xml = function(dom) {
      var options = _.defaults({ xmlMode: true }, this._options);
      return render(this, dom, options);
    };
    exports2.text = function(elems) {
      if (!elems) {
        elems = this.root();
      }
      var ret = "", len = elems.length, elem;
      for (var i = 0; i < len; i++) {
        elem = elems[i];
        if (elem.type === "text") ret += elem.data;
        else if (elem.children && elem.type !== "comment") {
          ret += exports2.text(elem.children);
        }
      }
      return ret;
    };
    exports2.parseHTML = function(data, context, keepScripts) {
      var parsed;
      if (!data || typeof data !== "string") {
        return null;
      }
      if (typeof context === "boolean") {
        keepScripts = context;
      }
      parsed = this.load(data);
      if (!keepScripts) {
        parsed("script").remove();
      }
      return parsed.root()[0].children.slice();
    };
    exports2.root = function() {
      return this(this._root);
    };
    exports2.contains = function(container, contained) {
      if (contained === container) {
        return false;
      }
      while (contained && contained !== contained.parent) {
        contained = contained.parent;
        if (contained === container) {
          return true;
        }
      }
      return false;
    };
  }
});

// node_modules/lodash.some/index.js
var require_lodash6 = __commonJS({
  "node_modules/lodash.some/index.js"(exports2, module2) {
    init_process_shim();
    var LARGE_ARRAY_SIZE = 200;
    var FUNC_ERROR_TEXT = "Expected a function";
    var HASH_UNDEFINED = "__lodash_hash_undefined__";
    var UNORDERED_COMPARE_FLAG = 1;
    var PARTIAL_COMPARE_FLAG = 2;
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var arrayTag = "[object Array]";
    var boolTag = "[object Boolean]";
    var dateTag = "[object Date]";
    var errorTag = "[object Error]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var mapTag = "[object Map]";
    var numberTag = "[object Number]";
    var objectTag = "[object Object]";
    var promiseTag = "[object Promise]";
    var regexpTag = "[object RegExp]";
    var setTag = "[object Set]";
    var stringTag = "[object String]";
    var symbolTag = "[object Symbol]";
    var weakMapTag = "[object WeakMap]";
    var arrayBufferTag = "[object ArrayBuffer]";
    var dataViewTag = "[object DataView]";
    var float32Tag = "[object Float32Array]";
    var float64Tag = "[object Float64Array]";
    var int8Tag = "[object Int8Array]";
    var int16Tag = "[object Int16Array]";
    var int32Tag = "[object Int32Array]";
    var uint8Tag = "[object Uint8Array]";
    var uint8ClampedTag = "[object Uint8ClampedArray]";
    var uint16Tag = "[object Uint16Array]";
    var uint32Tag = "[object Uint32Array]";
    var reIsDeepProp = /\.|\[(?:[^[\]]*|(["'])(?:(?!\1)[^\\]|\\.)*?\1)\]/;
    var reIsPlainProp = /^\w*$/;
    var reLeadingDot = /^\./;
    var rePropName = /[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g;
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reEscapeChar = /\\(\\)?/g;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var typedArrayTags = {};
    typedArrayTags[float32Tag] = typedArrayTags[float64Tag] = typedArrayTags[int8Tag] = typedArrayTags[int16Tag] = typedArrayTags[int32Tag] = typedArrayTags[uint8Tag] = typedArrayTags[uint8ClampedTag] = typedArrayTags[uint16Tag] = typedArrayTags[uint32Tag] = true;
    typedArrayTags[argsTag] = typedArrayTags[arrayTag] = typedArrayTags[arrayBufferTag] = typedArrayTags[boolTag] = typedArrayTags[dataViewTag] = typedArrayTags[dateTag] = typedArrayTags[errorTag] = typedArrayTags[funcTag] = typedArrayTags[mapTag] = typedArrayTags[numberTag] = typedArrayTags[objectTag] = typedArrayTags[regexpTag] = typedArrayTags[setTag] = typedArrayTags[stringTag] = typedArrayTags[weakMapTag] = false;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    var freeExports = typeof exports2 == "object" && exports2 && !exports2.nodeType && exports2;
    var freeModule = freeExports && typeof module2 == "object" && module2 && !module2.nodeType && module2;
    var moduleExports = freeModule && freeModule.exports === freeExports;
    var freeProcess = moduleExports && freeGlobal.process;
    var nodeUtil = (function() {
      try {
        return freeProcess && freeProcess.binding("util");
      } catch (e) {
      }
    })();
    var nodeIsTypedArray = nodeUtil && nodeUtil.isTypedArray;
    function arraySome(array, predicate) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (predicate(array[index], index, array)) {
          return true;
        }
      }
      return false;
    }
    function baseProperty(key) {
      return function(object) {
        return object == null ? void 0 : object[key];
      };
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function baseUnary(func) {
      return function(value) {
        return func(value);
      };
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function isHostObject(value) {
      var result = false;
      if (value != null && typeof value.toString != "function") {
        try {
          result = !!(value + "");
        } catch (e) {
        }
      }
      return result;
    }
    function mapToArray(map) {
      var index = -1, result = Array(map.size);
      map.forEach(function(value, key) {
        result[++index] = [key, value];
      });
      return result;
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    function setToArray(set) {
      var index = -1, result = Array(set.size);
      set.forEach(function(value) {
        result[++index] = value;
      });
      return result;
    }
    var arrayProto = Array.prototype;
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var Symbol2 = root.Symbol;
    var Uint8Array2 = root.Uint8Array;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var splice = arrayProto.splice;
    var nativeKeys = overArg(Object.keys, Object);
    var DataView = getNative(root, "DataView");
    var Map = getNative(root, "Map");
    var Promise2 = getNative(root, "Promise");
    var Set = getNative(root, "Set");
    var WeakMap = getNative(root, "WeakMap");
    var nativeCreate = getNative(Object, "create");
    var dataViewCtorString = toSource(DataView);
    var mapCtorString = toSource(Map);
    var promiseCtorString = toSource(Promise2);
    var setCtorString = toSource(Set);
    var weakMapCtorString = toSource(WeakMap);
    var symbolProto = Symbol2 ? Symbol2.prototype : void 0;
    var symbolValueOf = symbolProto ? symbolProto.valueOf : void 0;
    var symbolToString = symbolProto ? symbolProto.toString : void 0;
    function Hash(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function hashClear() {
      this.__data__ = nativeCreate ? nativeCreate(null) : {};
    }
    function hashDelete(key) {
      return this.has(key) && delete this.__data__[key];
    }
    function hashGet(key) {
      var data = this.__data__;
      if (nativeCreate) {
        var result = data[key];
        return result === HASH_UNDEFINED ? void 0 : result;
      }
      return hasOwnProperty2.call(data, key) ? data[key] : void 0;
    }
    function hashHas(key) {
      var data = this.__data__;
      return nativeCreate ? data[key] !== void 0 : hasOwnProperty2.call(data, key);
    }
    function hashSet(key, value) {
      var data = this.__data__;
      data[key] = nativeCreate && value === void 0 ? HASH_UNDEFINED : value;
      return this;
    }
    Hash.prototype.clear = hashClear;
    Hash.prototype["delete"] = hashDelete;
    Hash.prototype.get = hashGet;
    Hash.prototype.has = hashHas;
    Hash.prototype.set = hashSet;
    function ListCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function listCacheClear() {
      this.__data__ = [];
    }
    function listCacheDelete(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        return false;
      }
      var lastIndex = data.length - 1;
      if (index == lastIndex) {
        data.pop();
      } else {
        splice.call(data, index, 1);
      }
      return true;
    }
    function listCacheGet(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      return index < 0 ? void 0 : data[index][1];
    }
    function listCacheHas(key) {
      return assocIndexOf(this.__data__, key) > -1;
    }
    function listCacheSet(key, value) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        data.push([key, value]);
      } else {
        data[index][1] = value;
      }
      return this;
    }
    ListCache.prototype.clear = listCacheClear;
    ListCache.prototype["delete"] = listCacheDelete;
    ListCache.prototype.get = listCacheGet;
    ListCache.prototype.has = listCacheHas;
    ListCache.prototype.set = listCacheSet;
    function MapCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function mapCacheClear() {
      this.__data__ = {
        "hash": new Hash(),
        "map": new (Map || ListCache)(),
        "string": new Hash()
      };
    }
    function mapCacheDelete(key) {
      return getMapData(this, key)["delete"](key);
    }
    function mapCacheGet(key) {
      return getMapData(this, key).get(key);
    }
    function mapCacheHas(key) {
      return getMapData(this, key).has(key);
    }
    function mapCacheSet(key, value) {
      getMapData(this, key).set(key, value);
      return this;
    }
    MapCache.prototype.clear = mapCacheClear;
    MapCache.prototype["delete"] = mapCacheDelete;
    MapCache.prototype.get = mapCacheGet;
    MapCache.prototype.has = mapCacheHas;
    MapCache.prototype.set = mapCacheSet;
    function SetCache(values) {
      var index = -1, length = values ? values.length : 0;
      this.__data__ = new MapCache();
      while (++index < length) {
        this.add(values[index]);
      }
    }
    function setCacheAdd(value) {
      this.__data__.set(value, HASH_UNDEFINED);
      return this;
    }
    function setCacheHas(value) {
      return this.__data__.has(value);
    }
    SetCache.prototype.add = SetCache.prototype.push = setCacheAdd;
    SetCache.prototype.has = setCacheHas;
    function Stack(entries) {
      this.__data__ = new ListCache(entries);
    }
    function stackClear() {
      this.__data__ = new ListCache();
    }
    function stackDelete(key) {
      return this.__data__["delete"](key);
    }
    function stackGet(key) {
      return this.__data__.get(key);
    }
    function stackHas(key) {
      return this.__data__.has(key);
    }
    function stackSet(key, value) {
      var cache = this.__data__;
      if (cache instanceof ListCache) {
        var pairs = cache.__data__;
        if (!Map || pairs.length < LARGE_ARRAY_SIZE - 1) {
          pairs.push([key, value]);
          return this;
        }
        cache = this.__data__ = new MapCache(pairs);
      }
      cache.set(key, value);
      return this;
    }
    Stack.prototype.clear = stackClear;
    Stack.prototype["delete"] = stackDelete;
    Stack.prototype.get = stackGet;
    Stack.prototype.has = stackHas;
    Stack.prototype.set = stackSet;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assocIndexOf(array, key) {
      var length = array.length;
      while (length--) {
        if (eq(array[length][0], key)) {
          return length;
        }
      }
      return -1;
    }
    var baseEach = createBaseEach(baseForOwn);
    var baseFor = createBaseFor();
    function baseForOwn(object, iteratee) {
      return object && baseFor(object, iteratee, keys);
    }
    function baseGet(object, path) {
      path = isKey(path, object) ? [path] : castPath(path);
      var index = 0, length = path.length;
      while (object != null && index < length) {
        object = object[toKey(path[index++])];
      }
      return index && index == length ? object : void 0;
    }
    function baseGetTag(value) {
      return objectToString.call(value);
    }
    function baseHasIn(object, key) {
      return object != null && key in Object(object);
    }
    function baseIsEqual(value, other, customizer, bitmask, stack) {
      if (value === other) {
        return true;
      }
      if (value == null || other == null || !isObject(value) && !isObjectLike(other)) {
        return value !== value && other !== other;
      }
      return baseIsEqualDeep(value, other, baseIsEqual, customizer, bitmask, stack);
    }
    function baseIsEqualDeep(object, other, equalFunc, customizer, bitmask, stack) {
      var objIsArr = isArray(object), othIsArr = isArray(other), objTag = arrayTag, othTag = arrayTag;
      if (!objIsArr) {
        objTag = getTag(object);
        objTag = objTag == argsTag ? objectTag : objTag;
      }
      if (!othIsArr) {
        othTag = getTag(other);
        othTag = othTag == argsTag ? objectTag : othTag;
      }
      var objIsObj = objTag == objectTag && !isHostObject(object), othIsObj = othTag == objectTag && !isHostObject(other), isSameTag = objTag == othTag;
      if (isSameTag && !objIsObj) {
        stack || (stack = new Stack());
        return objIsArr || isTypedArray(object) ? equalArrays(object, other, equalFunc, customizer, bitmask, stack) : equalByTag(object, other, objTag, equalFunc, customizer, bitmask, stack);
      }
      if (!(bitmask & PARTIAL_COMPARE_FLAG)) {
        var objIsWrapped = objIsObj && hasOwnProperty2.call(object, "__wrapped__"), othIsWrapped = othIsObj && hasOwnProperty2.call(other, "__wrapped__");
        if (objIsWrapped || othIsWrapped) {
          var objUnwrapped = objIsWrapped ? object.value() : object, othUnwrapped = othIsWrapped ? other.value() : other;
          stack || (stack = new Stack());
          return equalFunc(objUnwrapped, othUnwrapped, customizer, bitmask, stack);
        }
      }
      if (!isSameTag) {
        return false;
      }
      stack || (stack = new Stack());
      return equalObjects(object, other, equalFunc, customizer, bitmask, stack);
    }
    function baseIsMatch(object, source, matchData, customizer) {
      var index = matchData.length, length = index, noCustomizer = !customizer;
      if (object == null) {
        return !length;
      }
      object = Object(object);
      while (index--) {
        var data = matchData[index];
        if (noCustomizer && data[2] ? data[1] !== object[data[0]] : !(data[0] in object)) {
          return false;
        }
      }
      while (++index < length) {
        data = matchData[index];
        var key = data[0], objValue = object[key], srcValue = data[1];
        if (noCustomizer && data[2]) {
          if (objValue === void 0 && !(key in object)) {
            return false;
          }
        } else {
          var stack = new Stack();
          if (customizer) {
            var result = customizer(objValue, srcValue, key, object, source, stack);
          }
          if (!(result === void 0 ? baseIsEqual(srcValue, objValue, customizer, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG, stack) : result)) {
            return false;
          }
        }
      }
      return true;
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) || isHostObject(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseIsTypedArray(value) {
      return isObjectLike(value) && isLength(value.length) && !!typedArrayTags[objectToString.call(value)];
    }
    function baseIteratee(value) {
      if (typeof value == "function") {
        return value;
      }
      if (value == null) {
        return identity;
      }
      if (typeof value == "object") {
        return isArray(value) ? baseMatchesProperty(value[0], value[1]) : baseMatches(value);
      }
      return property(value);
    }
    function baseKeys(object) {
      if (!isPrototype(object)) {
        return nativeKeys(object);
      }
      var result = [];
      for (var key in Object(object)) {
        if (hasOwnProperty2.call(object, key) && key != "constructor") {
          result.push(key);
        }
      }
      return result;
    }
    function baseMatches(source) {
      var matchData = getMatchData(source);
      if (matchData.length == 1 && matchData[0][2]) {
        return matchesStrictComparable(matchData[0][0], matchData[0][1]);
      }
      return function(object) {
        return object === source || baseIsMatch(object, source, matchData);
      };
    }
    function baseMatchesProperty(path, srcValue) {
      if (isKey(path) && isStrictComparable(srcValue)) {
        return matchesStrictComparable(toKey(path), srcValue);
      }
      return function(object) {
        var objValue = get(object, path);
        return objValue === void 0 && objValue === srcValue ? hasIn(object, path) : baseIsEqual(srcValue, objValue, void 0, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG);
      };
    }
    function basePropertyDeep(path) {
      return function(object) {
        return baseGet(object, path);
      };
    }
    function baseSome(collection, predicate) {
      var result;
      baseEach(collection, function(value, index, collection2) {
        result = predicate(value, index, collection2);
        return !result;
      });
      return !!result;
    }
    function baseToString(value) {
      if (typeof value == "string") {
        return value;
      }
      if (isSymbol(value)) {
        return symbolToString ? symbolToString.call(value) : "";
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function castPath(value) {
      return isArray(value) ? value : stringToPath(value);
    }
    function createBaseEach(eachFunc, fromRight) {
      return function(collection, iteratee) {
        if (collection == null) {
          return collection;
        }
        if (!isArrayLike(collection)) {
          return eachFunc(collection, iteratee);
        }
        var length = collection.length, index = fromRight ? length : -1, iterable = Object(collection);
        while (fromRight ? index-- : ++index < length) {
          if (iteratee(iterable[index], index, iterable) === false) {
            break;
          }
        }
        return collection;
      };
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function equalArrays(array, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, arrLength = array.length, othLength = other.length;
      if (arrLength != othLength && !(isPartial && othLength > arrLength)) {
        return false;
      }
      var stacked = stack.get(array);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var index = -1, result = true, seen = bitmask & UNORDERED_COMPARE_FLAG ? new SetCache() : void 0;
      stack.set(array, other);
      stack.set(other, array);
      while (++index < arrLength) {
        var arrValue = array[index], othValue = other[index];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, arrValue, index, other, array, stack) : customizer(arrValue, othValue, index, array, other, stack);
        }
        if (compared !== void 0) {
          if (compared) {
            continue;
          }
          result = false;
          break;
        }
        if (seen) {
          if (!arraySome(other, function(othValue2, othIndex) {
            if (!seen.has(othIndex) && (arrValue === othValue2 || equalFunc(arrValue, othValue2, customizer, bitmask, stack))) {
              return seen.add(othIndex);
            }
          })) {
            result = false;
            break;
          }
        } else if (!(arrValue === othValue || equalFunc(arrValue, othValue, customizer, bitmask, stack))) {
          result = false;
          break;
        }
      }
      stack["delete"](array);
      stack["delete"](other);
      return result;
    }
    function equalByTag(object, other, tag, equalFunc, customizer, bitmask, stack) {
      switch (tag) {
        case dataViewTag:
          if (object.byteLength != other.byteLength || object.byteOffset != other.byteOffset) {
            return false;
          }
          object = object.buffer;
          other = other.buffer;
        case arrayBufferTag:
          if (object.byteLength != other.byteLength || !equalFunc(new Uint8Array2(object), new Uint8Array2(other))) {
            return false;
          }
          return true;
        case boolTag:
        case dateTag:
        case numberTag:
          return eq(+object, +other);
        case errorTag:
          return object.name == other.name && object.message == other.message;
        case regexpTag:
        case stringTag:
          return object == other + "";
        case mapTag:
          var convert = mapToArray;
        case setTag:
          var isPartial = bitmask & PARTIAL_COMPARE_FLAG;
          convert || (convert = setToArray);
          if (object.size != other.size && !isPartial) {
            return false;
          }
          var stacked = stack.get(object);
          if (stacked) {
            return stacked == other;
          }
          bitmask |= UNORDERED_COMPARE_FLAG;
          stack.set(object, other);
          var result = equalArrays(convert(object), convert(other), equalFunc, customizer, bitmask, stack);
          stack["delete"](object);
          return result;
        case symbolTag:
          if (symbolValueOf) {
            return symbolValueOf.call(object) == symbolValueOf.call(other);
          }
      }
      return false;
    }
    function equalObjects(object, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, objProps = keys(object), objLength = objProps.length, othProps = keys(other), othLength = othProps.length;
      if (objLength != othLength && !isPartial) {
        return false;
      }
      var index = objLength;
      while (index--) {
        var key = objProps[index];
        if (!(isPartial ? key in other : hasOwnProperty2.call(other, key))) {
          return false;
        }
      }
      var stacked = stack.get(object);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var result = true;
      stack.set(object, other);
      stack.set(other, object);
      var skipCtor = isPartial;
      while (++index < objLength) {
        key = objProps[index];
        var objValue = object[key], othValue = other[key];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, objValue, key, other, object, stack) : customizer(objValue, othValue, key, object, other, stack);
        }
        if (!(compared === void 0 ? objValue === othValue || equalFunc(objValue, othValue, customizer, bitmask, stack) : compared)) {
          result = false;
          break;
        }
        skipCtor || (skipCtor = key == "constructor");
      }
      if (result && !skipCtor) {
        var objCtor = object.constructor, othCtor = other.constructor;
        if (objCtor != othCtor && ("constructor" in object && "constructor" in other) && !(typeof objCtor == "function" && objCtor instanceof objCtor && typeof othCtor == "function" && othCtor instanceof othCtor)) {
          result = false;
        }
      }
      stack["delete"](object);
      stack["delete"](other);
      return result;
    }
    function getMapData(map, key) {
      var data = map.__data__;
      return isKeyable(key) ? data[typeof key == "string" ? "string" : "hash"] : data.map;
    }
    function getMatchData(object) {
      var result = keys(object), length = result.length;
      while (length--) {
        var key = result[length], value = object[key];
        result[length] = [key, value, isStrictComparable(value)];
      }
      return result;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    var getTag = baseGetTag;
    if (DataView && getTag(new DataView(new ArrayBuffer(1))) != dataViewTag || Map && getTag(new Map()) != mapTag || Promise2 && getTag(Promise2.resolve()) != promiseTag || Set && getTag(new Set()) != setTag || WeakMap && getTag(new WeakMap()) != weakMapTag) {
      getTag = function(value) {
        var result = objectToString.call(value), Ctor = result == objectTag ? value.constructor : void 0, ctorString = Ctor ? toSource(Ctor) : void 0;
        if (ctorString) {
          switch (ctorString) {
            case dataViewCtorString:
              return dataViewTag;
            case mapCtorString:
              return mapTag;
            case promiseCtorString:
              return promiseTag;
            case setCtorString:
              return setTag;
            case weakMapCtorString:
              return weakMapTag;
          }
        }
        return result;
      };
    }
    function hasPath(object, path, hasFunc) {
      path = isKey(path, object) ? [path] : castPath(path);
      var result, index = -1, length = path.length;
      while (++index < length) {
        var key = toKey(path[index]);
        if (!(result = object != null && hasFunc(object, key))) {
          break;
        }
        object = object[key];
      }
      if (result) {
        return result;
      }
      var length = object ? object.length : 0;
      return !!length && isLength(length) && isIndex(key, length) && (isArray(object) || isArguments(object));
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isIterateeCall(value, index, object) {
      if (!isObject(object)) {
        return false;
      }
      var type = typeof index;
      if (type == "number" ? isArrayLike(object) && isIndex(index, object.length) : type == "string" && index in object) {
        return eq(object[index], value);
      }
      return false;
    }
    function isKey(value, object) {
      if (isArray(value)) {
        return false;
      }
      var type = typeof value;
      if (type == "number" || type == "symbol" || type == "boolean" || value == null || isSymbol(value)) {
        return true;
      }
      return reIsPlainProp.test(value) || !reIsDeepProp.test(value) || object != null && value in Object(object);
    }
    function isKeyable(value) {
      var type = typeof value;
      return type == "string" || type == "number" || type == "symbol" || type == "boolean" ? value !== "__proto__" : value === null;
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function isStrictComparable(value) {
      return value === value && !isObject(value);
    }
    function matchesStrictComparable(key, srcValue) {
      return function(object) {
        if (object == null) {
          return false;
        }
        return object[key] === srcValue && (srcValue !== void 0 || key in Object(object));
      };
    }
    var stringToPath = memoize(function(string) {
      string = toString(string);
      var result = [];
      if (reLeadingDot.test(string)) {
        result.push("");
      }
      string.replace(rePropName, function(match, number, quote, string2) {
        result.push(quote ? string2.replace(reEscapeChar, "$1") : number || match);
      });
      return result;
    });
    function toKey(value) {
      if (typeof value == "string" || isSymbol(value)) {
        return value;
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function some(collection, predicate, guard) {
      var func = isArray(collection) ? arraySome : baseSome;
      if (guard && isIterateeCall(collection, predicate, guard)) {
        predicate = void 0;
      }
      return func(collection, baseIteratee(predicate, 3));
    }
    function memoize(func, resolver) {
      if (typeof func != "function" || resolver && typeof resolver != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      var memoized = function() {
        var args = arguments, key = resolver ? resolver.apply(this, args) : args[0], cache = memoized.cache;
        if (cache.has(key)) {
          return cache.get(key);
        }
        var result = func.apply(this, args);
        memoized.cache = cache.set(key, result);
        return result;
      };
      memoized.cache = new (memoize.Cache || MapCache)();
      return memoized;
    }
    memoize.Cache = MapCache;
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    var isTypedArray = nodeIsTypedArray ? baseUnary(nodeIsTypedArray) : baseIsTypedArray;
    function toString(value) {
      return value == null ? "" : baseToString(value);
    }
    function get(object, path, defaultValue) {
      var result = object == null ? void 0 : baseGet(object, path);
      return result === void 0 ? defaultValue : result;
    }
    function hasIn(object, path) {
      return object != null && hasPath(object, path, baseHasIn);
    }
    function keys(object) {
      return isArrayLike(object) ? arrayLikeKeys(object) : baseKeys(object);
    }
    function identity(value) {
      return value;
    }
    function property(path) {
      return isKey(path) ? baseProperty(toKey(path)) : basePropertyDeep(path);
    }
    module2.exports = some;
  }
});

// node_modules/cheerio/lib/api/attributes.js
var require_attributes2 = __commonJS({
  "node_modules/cheerio/lib/api/attributes.js"(exports2) {
    init_process_shim();
    var $ = require_static();
    var utils = require_utils();
    var isTag = utils.isTag;
    var domEach = utils.domEach;
    var hasOwn = Object.prototype.hasOwnProperty;
    var camelCase = utils.camelCase;
    var cssCase = utils.cssCase;
    var rspace = /\s+/;
    var dataAttrPrefix = "data-";
    var _ = {
      forEach: require_lodash3(),
      extend: require_lodash(),
      some: require_lodash6()
    };
    var primitives = {
      null: null,
      true: true,
      false: false
    };
    var rboolean = /^(?:autofocus|autoplay|async|checked|controls|defer|disabled|hidden|loop|multiple|open|readonly|required|scoped|selected)$/i;
    var rbrace = /^(?:\{[\w\W]*\}|\[[\w\W]*\])$/;
    var getAttr = function(elem, name) {
      if (!elem || !isTag(elem)) return;
      if (!elem.attribs) {
        elem.attribs = {};
      }
      if (!name) {
        return elem.attribs;
      }
      if (hasOwn.call(elem.attribs, name)) {
        return rboolean.test(name) ? name : elem.attribs[name];
      }
      if (elem.name === "option" && name === "value") {
        return $.text(elem.children);
      }
      if (elem.name === "input" && (elem.attribs.type === "radio" || elem.attribs.type === "checkbox") && name === "value") {
        return "on";
      }
    };
    var setAttr = function(el, name, value) {
      if (value === null) {
        removeAttribute(el, name);
      } else {
        el.attribs[name] = value + "";
      }
    };
    exports2.attr = function(name, value) {
      if (typeof name === "object" || value !== void 0) {
        if (typeof value === "function") {
          return domEach(this, function(i, el) {
            setAttr(el, name, value.call(el, i, el.attribs[name]));
          });
        }
        return domEach(this, function(i, el) {
          if (!isTag(el)) return;
          if (typeof name === "object") {
            _.forEach(name, function(value2, name2) {
              setAttr(el, name2, value2);
            });
          } else {
            setAttr(el, name, value);
          }
        });
      }
      return getAttr(this[0], name);
    };
    var getProp = function(el, name) {
      if (!el || !isTag(el)) return;
      return el.hasOwnProperty(name) ? el[name] : rboolean.test(name) ? getAttr(el, name) !== void 0 : getAttr(el, name);
    };
    var setProp = function(el, name, value) {
      el[name] = rboolean.test(name) ? !!value : value;
    };
    exports2.prop = function(name, value) {
      var i = 0, property;
      if (typeof name === "string" && value === void 0) {
        switch (name) {
          case "style":
            property = this.css();
            _.forEach(property, function(v, p) {
              property[i++] = p;
            });
            property.length = i;
            break;
          case "tagName":
          case "nodeName":
            property = this[0].name.toUpperCase();
            break;
          default:
            property = getProp(this[0], name);
        }
        return property;
      }
      if (typeof name === "object" || value !== void 0) {
        if (typeof value === "function") {
          return domEach(this, function(i2, el) {
            setProp(el, name, value.call(el, i2, getProp(el, name)));
          });
        }
        return domEach(this, function(i2, el) {
          if (!isTag(el)) return;
          if (typeof name === "object") {
            _.forEach(name, function(val, name2) {
              setProp(el, name2, val);
            });
          } else {
            setProp(el, name, value);
          }
        });
      }
    };
    var setData = function(el, name, value) {
      if (!el.data) {
        el.data = {};
      }
      if (typeof name === "object") return _.extend(el.data, name);
      if (typeof name === "string" && value !== void 0) {
        el.data[name] = value;
      } else if (typeof name === "object") {
        _.extend(el.data, name);
      }
    };
    var readData = function(el, name) {
      var readAll = arguments.length === 1;
      var domNames, domName, jsNames, jsName, value, idx, length;
      if (readAll) {
        domNames = Object.keys(el.attribs).filter(function(attrName) {
          return attrName.slice(0, dataAttrPrefix.length) === dataAttrPrefix;
        });
        jsNames = domNames.map(function(domName2) {
          return camelCase(domName2.slice(dataAttrPrefix.length));
        });
      } else {
        domNames = [dataAttrPrefix + cssCase(name)];
        jsNames = [name];
      }
      for (idx = 0, length = domNames.length; idx < length; ++idx) {
        domName = domNames[idx];
        jsName = jsNames[idx];
        if (hasOwn.call(el.attribs, domName)) {
          value = el.attribs[domName];
          if (hasOwn.call(primitives, value)) {
            value = primitives[value];
          } else if (value === String(Number(value))) {
            value = Number(value);
          } else if (rbrace.test(value)) {
            try {
              value = JSON.parse(value);
            } catch (e) {
            }
          }
          el.data[jsName] = value;
        }
      }
      return readAll ? el.data : value;
    };
    exports2.data = function(name, value) {
      var elem = this[0];
      if (!elem || !isTag(elem)) return;
      if (!elem.data) {
        elem.data = {};
      }
      if (!name) {
        return readData(elem);
      }
      if (typeof name === "object" || value !== void 0) {
        domEach(this, function(i, el) {
          setData(el, name, value);
        });
        return this;
      } else if (hasOwn.call(elem.data, name)) {
        return elem.data[name];
      }
      return readData(elem, name);
    };
    exports2.val = function(value) {
      var querying = arguments.length === 0, element = this[0];
      if (!element) return;
      switch (element.name) {
        case "textarea":
          return this.text(value);
        case "input":
          switch (this.attr("type")) {
            case "radio":
              if (querying) {
                return this.attr("value");
              } else {
                this.attr("value", value);
                return this;
              }
              break;
            default:
              return this.attr("value", value);
          }
          return;
        case "select":
          var option = this.find("option:selected"), returnValue;
          if (option === void 0) return void 0;
          if (!querying) {
            if (!this.attr().hasOwnProperty("multiple") && typeof value == "object") {
              return this;
            }
            if (typeof value != "object") {
              value = [value];
            }
            this.find("option").removeAttr("selected");
            for (var i = 0; i < value.length; i++) {
              this.find('option[value="' + value[i] + '"]').attr("selected", "");
            }
            return this;
          }
          returnValue = option.attr("value");
          if (this.attr().hasOwnProperty("multiple")) {
            returnValue = [];
            domEach(option, function(i2, el) {
              returnValue.push(getAttr(el, "value"));
            });
          }
          return returnValue;
        case "option":
          if (!querying) {
            this.attr("value", value);
            return this;
          }
          return this.attr("value");
      }
    };
    var removeAttribute = function(elem, name) {
      if (!elem.attribs || !hasOwn.call(elem.attribs, name))
        return;
      delete elem.attribs[name];
    };
    exports2.removeAttr = function(name) {
      domEach(this, function(i, elem) {
        removeAttribute(elem, name);
      });
      return this;
    };
    exports2.hasClass = function(className) {
      return _.some(this, function(elem) {
        var attrs = elem.attribs, clazz = attrs && attrs["class"], idx = -1, end;
        if (clazz) {
          while ((idx = clazz.indexOf(className, idx + 1)) > -1) {
            end = idx + className.length;
            if ((idx === 0 || rspace.test(clazz[idx - 1])) && (end === clazz.length || rspace.test(clazz[end]))) {
              return true;
            }
          }
        }
      });
    };
    exports2.addClass = function(value) {
      if (typeof value === "function") {
        return domEach(this, function(i2, el) {
          var className2 = el.attribs["class"] || "";
          exports2.addClass.call([el], value.call(el, i2, className2));
        });
      }
      if (!value || typeof value !== "string") return this;
      var classNames = value.split(rspace), numElements = this.length;
      for (var i = 0; i < numElements; i++) {
        if (!isTag(this[i])) continue;
        var className = getAttr(this[i], "class"), numClasses, setClass;
        if (!className) {
          setAttr(this[i], "class", classNames.join(" ").trim());
        } else {
          setClass = " " + className + " ";
          numClasses = classNames.length;
          for (var j = 0; j < numClasses; j++) {
            var appendClass = classNames[j] + " ";
            if (setClass.indexOf(" " + appendClass) < 0)
              setClass += appendClass;
          }
          setAttr(this[i], "class", setClass.trim());
        }
      }
      return this;
    };
    var splitClass = function(className) {
      return className ? className.trim().split(rspace) : [];
    };
    exports2.removeClass = function(value) {
      var classes, numClasses, removeAll;
      if (typeof value === "function") {
        return domEach(this, function(i, el) {
          exports2.removeClass.call(
            [el],
            value.call(el, i, el.attribs["class"] || "")
          );
        });
      }
      classes = splitClass(value);
      numClasses = classes.length;
      removeAll = arguments.length === 0;
      return domEach(this, function(i, el) {
        if (!isTag(el)) return;
        if (removeAll) {
          el.attribs.class = "";
        } else {
          var elClasses = splitClass(el.attribs.class), index, changed;
          for (var j = 0; j < numClasses; j++) {
            index = elClasses.indexOf(classes[j]);
            if (index >= 0) {
              elClasses.splice(index, 1);
              changed = true;
              j--;
            }
          }
          if (changed) {
            el.attribs.class = elClasses.join(" ");
          }
        }
      });
    };
    exports2.toggleClass = function(value, stateVal) {
      if (typeof value === "function") {
        return domEach(this, function(i2, el) {
          exports2.toggleClass.call(
            [el],
            value.call(el, i2, el.attribs["class"] || "", stateVal),
            stateVal
          );
        });
      }
      if (!value || typeof value !== "string") return this;
      var classNames = value.split(rspace), numClasses = classNames.length, state = typeof stateVal === "boolean" ? stateVal ? 1 : -1 : 0, numElements = this.length, elementClasses, index;
      for (var i = 0; i < numElements; i++) {
        if (!isTag(this[i])) continue;
        elementClasses = splitClass(this[i].attribs.class);
        for (var j = 0; j < numClasses; j++) {
          index = elementClasses.indexOf(classNames[j]);
          if (state >= 0 && index < 0) {
            elementClasses.push(classNames[j]);
          } else if (state <= 0 && index >= 0) {
            elementClasses.splice(index, 1);
          }
        }
        this[i].attribs.class = elementClasses.join(" ");
      }
      return this;
    };
    exports2.is = function(selector) {
      if (selector) {
        return this.filter(selector).length > 0;
      }
      return false;
    };
  }
});

// node_modules/lodash.reject/index.js
var require_lodash7 = __commonJS({
  "node_modules/lodash.reject/index.js"(exports2, module2) {
    init_process_shim();
    var LARGE_ARRAY_SIZE = 200;
    var FUNC_ERROR_TEXT = "Expected a function";
    var HASH_UNDEFINED = "__lodash_hash_undefined__";
    var UNORDERED_COMPARE_FLAG = 1;
    var PARTIAL_COMPARE_FLAG = 2;
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var arrayTag = "[object Array]";
    var boolTag = "[object Boolean]";
    var dateTag = "[object Date]";
    var errorTag = "[object Error]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var mapTag = "[object Map]";
    var numberTag = "[object Number]";
    var objectTag = "[object Object]";
    var promiseTag = "[object Promise]";
    var regexpTag = "[object RegExp]";
    var setTag = "[object Set]";
    var stringTag = "[object String]";
    var symbolTag = "[object Symbol]";
    var weakMapTag = "[object WeakMap]";
    var arrayBufferTag = "[object ArrayBuffer]";
    var dataViewTag = "[object DataView]";
    var float32Tag = "[object Float32Array]";
    var float64Tag = "[object Float64Array]";
    var int8Tag = "[object Int8Array]";
    var int16Tag = "[object Int16Array]";
    var int32Tag = "[object Int32Array]";
    var uint8Tag = "[object Uint8Array]";
    var uint8ClampedTag = "[object Uint8ClampedArray]";
    var uint16Tag = "[object Uint16Array]";
    var uint32Tag = "[object Uint32Array]";
    var reIsDeepProp = /\.|\[(?:[^[\]]*|(["'])(?:(?!\1)[^\\]|\\.)*?\1)\]/;
    var reIsPlainProp = /^\w*$/;
    var reLeadingDot = /^\./;
    var rePropName = /[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g;
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reEscapeChar = /\\(\\)?/g;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var typedArrayTags = {};
    typedArrayTags[float32Tag] = typedArrayTags[float64Tag] = typedArrayTags[int8Tag] = typedArrayTags[int16Tag] = typedArrayTags[int32Tag] = typedArrayTags[uint8Tag] = typedArrayTags[uint8ClampedTag] = typedArrayTags[uint16Tag] = typedArrayTags[uint32Tag] = true;
    typedArrayTags[argsTag] = typedArrayTags[arrayTag] = typedArrayTags[arrayBufferTag] = typedArrayTags[boolTag] = typedArrayTags[dataViewTag] = typedArrayTags[dateTag] = typedArrayTags[errorTag] = typedArrayTags[funcTag] = typedArrayTags[mapTag] = typedArrayTags[numberTag] = typedArrayTags[objectTag] = typedArrayTags[regexpTag] = typedArrayTags[setTag] = typedArrayTags[stringTag] = typedArrayTags[weakMapTag] = false;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    var freeExports = typeof exports2 == "object" && exports2 && !exports2.nodeType && exports2;
    var freeModule = freeExports && typeof module2 == "object" && module2 && !module2.nodeType && module2;
    var moduleExports = freeModule && freeModule.exports === freeExports;
    var freeProcess = moduleExports && freeGlobal.process;
    var nodeUtil = (function() {
      try {
        return freeProcess && freeProcess.binding("util");
      } catch (e) {
      }
    })();
    var nodeIsTypedArray = nodeUtil && nodeUtil.isTypedArray;
    function arrayFilter(array, predicate) {
      var index = -1, length = array ? array.length : 0, resIndex = 0, result = [];
      while (++index < length) {
        var value = array[index];
        if (predicate(value, index, array)) {
          result[resIndex++] = value;
        }
      }
      return result;
    }
    function arraySome(array, predicate) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (predicate(array[index], index, array)) {
          return true;
        }
      }
      return false;
    }
    function baseProperty(key) {
      return function(object) {
        return object == null ? void 0 : object[key];
      };
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function baseUnary(func) {
      return function(value) {
        return func(value);
      };
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function isHostObject(value) {
      var result = false;
      if (value != null && typeof value.toString != "function") {
        try {
          result = !!(value + "");
        } catch (e) {
        }
      }
      return result;
    }
    function mapToArray(map) {
      var index = -1, result = Array(map.size);
      map.forEach(function(value, key) {
        result[++index] = [key, value];
      });
      return result;
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    function setToArray(set) {
      var index = -1, result = Array(set.size);
      set.forEach(function(value) {
        result[++index] = value;
      });
      return result;
    }
    var arrayProto = Array.prototype;
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var Symbol2 = root.Symbol;
    var Uint8Array2 = root.Uint8Array;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var splice = arrayProto.splice;
    var nativeKeys = overArg(Object.keys, Object);
    var DataView = getNative(root, "DataView");
    var Map = getNative(root, "Map");
    var Promise2 = getNative(root, "Promise");
    var Set = getNative(root, "Set");
    var WeakMap = getNative(root, "WeakMap");
    var nativeCreate = getNative(Object, "create");
    var dataViewCtorString = toSource(DataView);
    var mapCtorString = toSource(Map);
    var promiseCtorString = toSource(Promise2);
    var setCtorString = toSource(Set);
    var weakMapCtorString = toSource(WeakMap);
    var symbolProto = Symbol2 ? Symbol2.prototype : void 0;
    var symbolValueOf = symbolProto ? symbolProto.valueOf : void 0;
    var symbolToString = symbolProto ? symbolProto.toString : void 0;
    function Hash(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function hashClear() {
      this.__data__ = nativeCreate ? nativeCreate(null) : {};
    }
    function hashDelete(key) {
      return this.has(key) && delete this.__data__[key];
    }
    function hashGet(key) {
      var data = this.__data__;
      if (nativeCreate) {
        var result = data[key];
        return result === HASH_UNDEFINED ? void 0 : result;
      }
      return hasOwnProperty2.call(data, key) ? data[key] : void 0;
    }
    function hashHas(key) {
      var data = this.__data__;
      return nativeCreate ? data[key] !== void 0 : hasOwnProperty2.call(data, key);
    }
    function hashSet(key, value) {
      var data = this.__data__;
      data[key] = nativeCreate && value === void 0 ? HASH_UNDEFINED : value;
      return this;
    }
    Hash.prototype.clear = hashClear;
    Hash.prototype["delete"] = hashDelete;
    Hash.prototype.get = hashGet;
    Hash.prototype.has = hashHas;
    Hash.prototype.set = hashSet;
    function ListCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function listCacheClear() {
      this.__data__ = [];
    }
    function listCacheDelete(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        return false;
      }
      var lastIndex = data.length - 1;
      if (index == lastIndex) {
        data.pop();
      } else {
        splice.call(data, index, 1);
      }
      return true;
    }
    function listCacheGet(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      return index < 0 ? void 0 : data[index][1];
    }
    function listCacheHas(key) {
      return assocIndexOf(this.__data__, key) > -1;
    }
    function listCacheSet(key, value) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        data.push([key, value]);
      } else {
        data[index][1] = value;
      }
      return this;
    }
    ListCache.prototype.clear = listCacheClear;
    ListCache.prototype["delete"] = listCacheDelete;
    ListCache.prototype.get = listCacheGet;
    ListCache.prototype.has = listCacheHas;
    ListCache.prototype.set = listCacheSet;
    function MapCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function mapCacheClear() {
      this.__data__ = {
        "hash": new Hash(),
        "map": new (Map || ListCache)(),
        "string": new Hash()
      };
    }
    function mapCacheDelete(key) {
      return getMapData(this, key)["delete"](key);
    }
    function mapCacheGet(key) {
      return getMapData(this, key).get(key);
    }
    function mapCacheHas(key) {
      return getMapData(this, key).has(key);
    }
    function mapCacheSet(key, value) {
      getMapData(this, key).set(key, value);
      return this;
    }
    MapCache.prototype.clear = mapCacheClear;
    MapCache.prototype["delete"] = mapCacheDelete;
    MapCache.prototype.get = mapCacheGet;
    MapCache.prototype.has = mapCacheHas;
    MapCache.prototype.set = mapCacheSet;
    function SetCache(values) {
      var index = -1, length = values ? values.length : 0;
      this.__data__ = new MapCache();
      while (++index < length) {
        this.add(values[index]);
      }
    }
    function setCacheAdd(value) {
      this.__data__.set(value, HASH_UNDEFINED);
      return this;
    }
    function setCacheHas(value) {
      return this.__data__.has(value);
    }
    SetCache.prototype.add = SetCache.prototype.push = setCacheAdd;
    SetCache.prototype.has = setCacheHas;
    function Stack(entries) {
      this.__data__ = new ListCache(entries);
    }
    function stackClear() {
      this.__data__ = new ListCache();
    }
    function stackDelete(key) {
      return this.__data__["delete"](key);
    }
    function stackGet(key) {
      return this.__data__.get(key);
    }
    function stackHas(key) {
      return this.__data__.has(key);
    }
    function stackSet(key, value) {
      var cache = this.__data__;
      if (cache instanceof ListCache) {
        var pairs = cache.__data__;
        if (!Map || pairs.length < LARGE_ARRAY_SIZE - 1) {
          pairs.push([key, value]);
          return this;
        }
        cache = this.__data__ = new MapCache(pairs);
      }
      cache.set(key, value);
      return this;
    }
    Stack.prototype.clear = stackClear;
    Stack.prototype["delete"] = stackDelete;
    Stack.prototype.get = stackGet;
    Stack.prototype.has = stackHas;
    Stack.prototype.set = stackSet;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assocIndexOf(array, key) {
      var length = array.length;
      while (length--) {
        if (eq(array[length][0], key)) {
          return length;
        }
      }
      return -1;
    }
    var baseEach = createBaseEach(baseForOwn);
    function baseFilter(collection, predicate) {
      var result = [];
      baseEach(collection, function(value, index, collection2) {
        if (predicate(value, index, collection2)) {
          result.push(value);
        }
      });
      return result;
    }
    var baseFor = createBaseFor();
    function baseForOwn(object, iteratee) {
      return object && baseFor(object, iteratee, keys);
    }
    function baseGet(object, path) {
      path = isKey(path, object) ? [path] : castPath(path);
      var index = 0, length = path.length;
      while (object != null && index < length) {
        object = object[toKey(path[index++])];
      }
      return index && index == length ? object : void 0;
    }
    function baseGetTag(value) {
      return objectToString.call(value);
    }
    function baseHasIn(object, key) {
      return object != null && key in Object(object);
    }
    function baseIsEqual(value, other, customizer, bitmask, stack) {
      if (value === other) {
        return true;
      }
      if (value == null || other == null || !isObject(value) && !isObjectLike(other)) {
        return value !== value && other !== other;
      }
      return baseIsEqualDeep(value, other, baseIsEqual, customizer, bitmask, stack);
    }
    function baseIsEqualDeep(object, other, equalFunc, customizer, bitmask, stack) {
      var objIsArr = isArray(object), othIsArr = isArray(other), objTag = arrayTag, othTag = arrayTag;
      if (!objIsArr) {
        objTag = getTag(object);
        objTag = objTag == argsTag ? objectTag : objTag;
      }
      if (!othIsArr) {
        othTag = getTag(other);
        othTag = othTag == argsTag ? objectTag : othTag;
      }
      var objIsObj = objTag == objectTag && !isHostObject(object), othIsObj = othTag == objectTag && !isHostObject(other), isSameTag = objTag == othTag;
      if (isSameTag && !objIsObj) {
        stack || (stack = new Stack());
        return objIsArr || isTypedArray(object) ? equalArrays(object, other, equalFunc, customizer, bitmask, stack) : equalByTag(object, other, objTag, equalFunc, customizer, bitmask, stack);
      }
      if (!(bitmask & PARTIAL_COMPARE_FLAG)) {
        var objIsWrapped = objIsObj && hasOwnProperty2.call(object, "__wrapped__"), othIsWrapped = othIsObj && hasOwnProperty2.call(other, "__wrapped__");
        if (objIsWrapped || othIsWrapped) {
          var objUnwrapped = objIsWrapped ? object.value() : object, othUnwrapped = othIsWrapped ? other.value() : other;
          stack || (stack = new Stack());
          return equalFunc(objUnwrapped, othUnwrapped, customizer, bitmask, stack);
        }
      }
      if (!isSameTag) {
        return false;
      }
      stack || (stack = new Stack());
      return equalObjects(object, other, equalFunc, customizer, bitmask, stack);
    }
    function baseIsMatch(object, source, matchData, customizer) {
      var index = matchData.length, length = index, noCustomizer = !customizer;
      if (object == null) {
        return !length;
      }
      object = Object(object);
      while (index--) {
        var data = matchData[index];
        if (noCustomizer && data[2] ? data[1] !== object[data[0]] : !(data[0] in object)) {
          return false;
        }
      }
      while (++index < length) {
        data = matchData[index];
        var key = data[0], objValue = object[key], srcValue = data[1];
        if (noCustomizer && data[2]) {
          if (objValue === void 0 && !(key in object)) {
            return false;
          }
        } else {
          var stack = new Stack();
          if (customizer) {
            var result = customizer(objValue, srcValue, key, object, source, stack);
          }
          if (!(result === void 0 ? baseIsEqual(srcValue, objValue, customizer, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG, stack) : result)) {
            return false;
          }
        }
      }
      return true;
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) || isHostObject(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseIsTypedArray(value) {
      return isObjectLike(value) && isLength(value.length) && !!typedArrayTags[objectToString.call(value)];
    }
    function baseIteratee(value) {
      if (typeof value == "function") {
        return value;
      }
      if (value == null) {
        return identity;
      }
      if (typeof value == "object") {
        return isArray(value) ? baseMatchesProperty(value[0], value[1]) : baseMatches(value);
      }
      return property(value);
    }
    function baseKeys(object) {
      if (!isPrototype(object)) {
        return nativeKeys(object);
      }
      var result = [];
      for (var key in Object(object)) {
        if (hasOwnProperty2.call(object, key) && key != "constructor") {
          result.push(key);
        }
      }
      return result;
    }
    function baseMatches(source) {
      var matchData = getMatchData(source);
      if (matchData.length == 1 && matchData[0][2]) {
        return matchesStrictComparable(matchData[0][0], matchData[0][1]);
      }
      return function(object) {
        return object === source || baseIsMatch(object, source, matchData);
      };
    }
    function baseMatchesProperty(path, srcValue) {
      if (isKey(path) && isStrictComparable(srcValue)) {
        return matchesStrictComparable(toKey(path), srcValue);
      }
      return function(object) {
        var objValue = get(object, path);
        return objValue === void 0 && objValue === srcValue ? hasIn(object, path) : baseIsEqual(srcValue, objValue, void 0, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG);
      };
    }
    function basePropertyDeep(path) {
      return function(object) {
        return baseGet(object, path);
      };
    }
    function baseToString(value) {
      if (typeof value == "string") {
        return value;
      }
      if (isSymbol(value)) {
        return symbolToString ? symbolToString.call(value) : "";
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function castPath(value) {
      return isArray(value) ? value : stringToPath(value);
    }
    function createBaseEach(eachFunc, fromRight) {
      return function(collection, iteratee) {
        if (collection == null) {
          return collection;
        }
        if (!isArrayLike(collection)) {
          return eachFunc(collection, iteratee);
        }
        var length = collection.length, index = fromRight ? length : -1, iterable = Object(collection);
        while (fromRight ? index-- : ++index < length) {
          if (iteratee(iterable[index], index, iterable) === false) {
            break;
          }
        }
        return collection;
      };
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function equalArrays(array, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, arrLength = array.length, othLength = other.length;
      if (arrLength != othLength && !(isPartial && othLength > arrLength)) {
        return false;
      }
      var stacked = stack.get(array);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var index = -1, result = true, seen = bitmask & UNORDERED_COMPARE_FLAG ? new SetCache() : void 0;
      stack.set(array, other);
      stack.set(other, array);
      while (++index < arrLength) {
        var arrValue = array[index], othValue = other[index];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, arrValue, index, other, array, stack) : customizer(arrValue, othValue, index, array, other, stack);
        }
        if (compared !== void 0) {
          if (compared) {
            continue;
          }
          result = false;
          break;
        }
        if (seen) {
          if (!arraySome(other, function(othValue2, othIndex) {
            if (!seen.has(othIndex) && (arrValue === othValue2 || equalFunc(arrValue, othValue2, customizer, bitmask, stack))) {
              return seen.add(othIndex);
            }
          })) {
            result = false;
            break;
          }
        } else if (!(arrValue === othValue || equalFunc(arrValue, othValue, customizer, bitmask, stack))) {
          result = false;
          break;
        }
      }
      stack["delete"](array);
      stack["delete"](other);
      return result;
    }
    function equalByTag(object, other, tag, equalFunc, customizer, bitmask, stack) {
      switch (tag) {
        case dataViewTag:
          if (object.byteLength != other.byteLength || object.byteOffset != other.byteOffset) {
            return false;
          }
          object = object.buffer;
          other = other.buffer;
        case arrayBufferTag:
          if (object.byteLength != other.byteLength || !equalFunc(new Uint8Array2(object), new Uint8Array2(other))) {
            return false;
          }
          return true;
        case boolTag:
        case dateTag:
        case numberTag:
          return eq(+object, +other);
        case errorTag:
          return object.name == other.name && object.message == other.message;
        case regexpTag:
        case stringTag:
          return object == other + "";
        case mapTag:
          var convert = mapToArray;
        case setTag:
          var isPartial = bitmask & PARTIAL_COMPARE_FLAG;
          convert || (convert = setToArray);
          if (object.size != other.size && !isPartial) {
            return false;
          }
          var stacked = stack.get(object);
          if (stacked) {
            return stacked == other;
          }
          bitmask |= UNORDERED_COMPARE_FLAG;
          stack.set(object, other);
          var result = equalArrays(convert(object), convert(other), equalFunc, customizer, bitmask, stack);
          stack["delete"](object);
          return result;
        case symbolTag:
          if (symbolValueOf) {
            return symbolValueOf.call(object) == symbolValueOf.call(other);
          }
      }
      return false;
    }
    function equalObjects(object, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, objProps = keys(object), objLength = objProps.length, othProps = keys(other), othLength = othProps.length;
      if (objLength != othLength && !isPartial) {
        return false;
      }
      var index = objLength;
      while (index--) {
        var key = objProps[index];
        if (!(isPartial ? key in other : hasOwnProperty2.call(other, key))) {
          return false;
        }
      }
      var stacked = stack.get(object);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var result = true;
      stack.set(object, other);
      stack.set(other, object);
      var skipCtor = isPartial;
      while (++index < objLength) {
        key = objProps[index];
        var objValue = object[key], othValue = other[key];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, objValue, key, other, object, stack) : customizer(objValue, othValue, key, object, other, stack);
        }
        if (!(compared === void 0 ? objValue === othValue || equalFunc(objValue, othValue, customizer, bitmask, stack) : compared)) {
          result = false;
          break;
        }
        skipCtor || (skipCtor = key == "constructor");
      }
      if (result && !skipCtor) {
        var objCtor = object.constructor, othCtor = other.constructor;
        if (objCtor != othCtor && ("constructor" in object && "constructor" in other) && !(typeof objCtor == "function" && objCtor instanceof objCtor && typeof othCtor == "function" && othCtor instanceof othCtor)) {
          result = false;
        }
      }
      stack["delete"](object);
      stack["delete"](other);
      return result;
    }
    function getMapData(map, key) {
      var data = map.__data__;
      return isKeyable(key) ? data[typeof key == "string" ? "string" : "hash"] : data.map;
    }
    function getMatchData(object) {
      var result = keys(object), length = result.length;
      while (length--) {
        var key = result[length], value = object[key];
        result[length] = [key, value, isStrictComparable(value)];
      }
      return result;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    var getTag = baseGetTag;
    if (DataView && getTag(new DataView(new ArrayBuffer(1))) != dataViewTag || Map && getTag(new Map()) != mapTag || Promise2 && getTag(Promise2.resolve()) != promiseTag || Set && getTag(new Set()) != setTag || WeakMap && getTag(new WeakMap()) != weakMapTag) {
      getTag = function(value) {
        var result = objectToString.call(value), Ctor = result == objectTag ? value.constructor : void 0, ctorString = Ctor ? toSource(Ctor) : void 0;
        if (ctorString) {
          switch (ctorString) {
            case dataViewCtorString:
              return dataViewTag;
            case mapCtorString:
              return mapTag;
            case promiseCtorString:
              return promiseTag;
            case setCtorString:
              return setTag;
            case weakMapCtorString:
              return weakMapTag;
          }
        }
        return result;
      };
    }
    function hasPath(object, path, hasFunc) {
      path = isKey(path, object) ? [path] : castPath(path);
      var result, index = -1, length = path.length;
      while (++index < length) {
        var key = toKey(path[index]);
        if (!(result = object != null && hasFunc(object, key))) {
          break;
        }
        object = object[key];
      }
      if (result) {
        return result;
      }
      var length = object ? object.length : 0;
      return !!length && isLength(length) && isIndex(key, length) && (isArray(object) || isArguments(object));
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isKey(value, object) {
      if (isArray(value)) {
        return false;
      }
      var type = typeof value;
      if (type == "number" || type == "symbol" || type == "boolean" || value == null || isSymbol(value)) {
        return true;
      }
      return reIsPlainProp.test(value) || !reIsDeepProp.test(value) || object != null && value in Object(object);
    }
    function isKeyable(value) {
      var type = typeof value;
      return type == "string" || type == "number" || type == "symbol" || type == "boolean" ? value !== "__proto__" : value === null;
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function isStrictComparable(value) {
      return value === value && !isObject(value);
    }
    function matchesStrictComparable(key, srcValue) {
      return function(object) {
        if (object == null) {
          return false;
        }
        return object[key] === srcValue && (srcValue !== void 0 || key in Object(object));
      };
    }
    var stringToPath = memoize(function(string) {
      string = toString(string);
      var result = [];
      if (reLeadingDot.test(string)) {
        result.push("");
      }
      string.replace(rePropName, function(match, number, quote, string2) {
        result.push(quote ? string2.replace(reEscapeChar, "$1") : number || match);
      });
      return result;
    });
    function toKey(value) {
      if (typeof value == "string" || isSymbol(value)) {
        return value;
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function reject(collection, predicate) {
      var func = isArray(collection) ? arrayFilter : baseFilter;
      return func(collection, negate(baseIteratee(predicate, 3)));
    }
    function memoize(func, resolver) {
      if (typeof func != "function" || resolver && typeof resolver != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      var memoized = function() {
        var args = arguments, key = resolver ? resolver.apply(this, args) : args[0], cache = memoized.cache;
        if (cache.has(key)) {
          return cache.get(key);
        }
        var result = func.apply(this, args);
        memoized.cache = cache.set(key, result);
        return result;
      };
      memoized.cache = new (memoize.Cache || MapCache)();
      return memoized;
    }
    memoize.Cache = MapCache;
    function negate(predicate) {
      if (typeof predicate != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      return function() {
        var args = arguments;
        switch (args.length) {
          case 0:
            return !predicate.call(this);
          case 1:
            return !predicate.call(this, args[0]);
          case 2:
            return !predicate.call(this, args[0], args[1]);
          case 3:
            return !predicate.call(this, args[0], args[1], args[2]);
        }
        return !predicate.apply(this, args);
      };
    }
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    var isTypedArray = nodeIsTypedArray ? baseUnary(nodeIsTypedArray) : baseIsTypedArray;
    function toString(value) {
      return value == null ? "" : baseToString(value);
    }
    function get(object, path, defaultValue) {
      var result = object == null ? void 0 : baseGet(object, path);
      return result === void 0 ? defaultValue : result;
    }
    function hasIn(object, path) {
      return object != null && hasPath(object, path, baseHasIn);
    }
    function keys(object) {
      return isArrayLike(object) ? arrayLikeKeys(object) : baseKeys(object);
    }
    function identity(value) {
      return value;
    }
    function property(path) {
      return isKey(path) ? baseProperty(toKey(path)) : basePropertyDeep(path);
    }
    module2.exports = reject;
  }
});

// node_modules/lodash.filter/index.js
var require_lodash8 = __commonJS({
  "node_modules/lodash.filter/index.js"(exports2, module2) {
    init_process_shim();
    var LARGE_ARRAY_SIZE = 200;
    var FUNC_ERROR_TEXT = "Expected a function";
    var HASH_UNDEFINED = "__lodash_hash_undefined__";
    var UNORDERED_COMPARE_FLAG = 1;
    var PARTIAL_COMPARE_FLAG = 2;
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var arrayTag = "[object Array]";
    var boolTag = "[object Boolean]";
    var dateTag = "[object Date]";
    var errorTag = "[object Error]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var mapTag = "[object Map]";
    var numberTag = "[object Number]";
    var objectTag = "[object Object]";
    var promiseTag = "[object Promise]";
    var regexpTag = "[object RegExp]";
    var setTag = "[object Set]";
    var stringTag = "[object String]";
    var symbolTag = "[object Symbol]";
    var weakMapTag = "[object WeakMap]";
    var arrayBufferTag = "[object ArrayBuffer]";
    var dataViewTag = "[object DataView]";
    var float32Tag = "[object Float32Array]";
    var float64Tag = "[object Float64Array]";
    var int8Tag = "[object Int8Array]";
    var int16Tag = "[object Int16Array]";
    var int32Tag = "[object Int32Array]";
    var uint8Tag = "[object Uint8Array]";
    var uint8ClampedTag = "[object Uint8ClampedArray]";
    var uint16Tag = "[object Uint16Array]";
    var uint32Tag = "[object Uint32Array]";
    var reIsDeepProp = /\.|\[(?:[^[\]]*|(["'])(?:(?!\1)[^\\]|\\.)*?\1)\]/;
    var reIsPlainProp = /^\w*$/;
    var reLeadingDot = /^\./;
    var rePropName = /[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g;
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reEscapeChar = /\\(\\)?/g;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var typedArrayTags = {};
    typedArrayTags[float32Tag] = typedArrayTags[float64Tag] = typedArrayTags[int8Tag] = typedArrayTags[int16Tag] = typedArrayTags[int32Tag] = typedArrayTags[uint8Tag] = typedArrayTags[uint8ClampedTag] = typedArrayTags[uint16Tag] = typedArrayTags[uint32Tag] = true;
    typedArrayTags[argsTag] = typedArrayTags[arrayTag] = typedArrayTags[arrayBufferTag] = typedArrayTags[boolTag] = typedArrayTags[dataViewTag] = typedArrayTags[dateTag] = typedArrayTags[errorTag] = typedArrayTags[funcTag] = typedArrayTags[mapTag] = typedArrayTags[numberTag] = typedArrayTags[objectTag] = typedArrayTags[regexpTag] = typedArrayTags[setTag] = typedArrayTags[stringTag] = typedArrayTags[weakMapTag] = false;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    var freeExports = typeof exports2 == "object" && exports2 && !exports2.nodeType && exports2;
    var freeModule = freeExports && typeof module2 == "object" && module2 && !module2.nodeType && module2;
    var moduleExports = freeModule && freeModule.exports === freeExports;
    var freeProcess = moduleExports && freeGlobal.process;
    var nodeUtil = (function() {
      try {
        return freeProcess && freeProcess.binding("util");
      } catch (e) {
      }
    })();
    var nodeIsTypedArray = nodeUtil && nodeUtil.isTypedArray;
    function arrayFilter(array, predicate) {
      var index = -1, length = array ? array.length : 0, resIndex = 0, result = [];
      while (++index < length) {
        var value = array[index];
        if (predicate(value, index, array)) {
          result[resIndex++] = value;
        }
      }
      return result;
    }
    function arraySome(array, predicate) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (predicate(array[index], index, array)) {
          return true;
        }
      }
      return false;
    }
    function baseProperty(key) {
      return function(object) {
        return object == null ? void 0 : object[key];
      };
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function baseUnary(func) {
      return function(value) {
        return func(value);
      };
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function isHostObject(value) {
      var result = false;
      if (value != null && typeof value.toString != "function") {
        try {
          result = !!(value + "");
        } catch (e) {
        }
      }
      return result;
    }
    function mapToArray(map) {
      var index = -1, result = Array(map.size);
      map.forEach(function(value, key) {
        result[++index] = [key, value];
      });
      return result;
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    function setToArray(set) {
      var index = -1, result = Array(set.size);
      set.forEach(function(value) {
        result[++index] = value;
      });
      return result;
    }
    var arrayProto = Array.prototype;
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var Symbol2 = root.Symbol;
    var Uint8Array2 = root.Uint8Array;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var splice = arrayProto.splice;
    var nativeKeys = overArg(Object.keys, Object);
    var DataView = getNative(root, "DataView");
    var Map = getNative(root, "Map");
    var Promise2 = getNative(root, "Promise");
    var Set = getNative(root, "Set");
    var WeakMap = getNative(root, "WeakMap");
    var nativeCreate = getNative(Object, "create");
    var dataViewCtorString = toSource(DataView);
    var mapCtorString = toSource(Map);
    var promiseCtorString = toSource(Promise2);
    var setCtorString = toSource(Set);
    var weakMapCtorString = toSource(WeakMap);
    var symbolProto = Symbol2 ? Symbol2.prototype : void 0;
    var symbolValueOf = symbolProto ? symbolProto.valueOf : void 0;
    var symbolToString = symbolProto ? symbolProto.toString : void 0;
    function Hash(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function hashClear() {
      this.__data__ = nativeCreate ? nativeCreate(null) : {};
    }
    function hashDelete(key) {
      return this.has(key) && delete this.__data__[key];
    }
    function hashGet(key) {
      var data = this.__data__;
      if (nativeCreate) {
        var result = data[key];
        return result === HASH_UNDEFINED ? void 0 : result;
      }
      return hasOwnProperty2.call(data, key) ? data[key] : void 0;
    }
    function hashHas(key) {
      var data = this.__data__;
      return nativeCreate ? data[key] !== void 0 : hasOwnProperty2.call(data, key);
    }
    function hashSet(key, value) {
      var data = this.__data__;
      data[key] = nativeCreate && value === void 0 ? HASH_UNDEFINED : value;
      return this;
    }
    Hash.prototype.clear = hashClear;
    Hash.prototype["delete"] = hashDelete;
    Hash.prototype.get = hashGet;
    Hash.prototype.has = hashHas;
    Hash.prototype.set = hashSet;
    function ListCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function listCacheClear() {
      this.__data__ = [];
    }
    function listCacheDelete(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        return false;
      }
      var lastIndex = data.length - 1;
      if (index == lastIndex) {
        data.pop();
      } else {
        splice.call(data, index, 1);
      }
      return true;
    }
    function listCacheGet(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      return index < 0 ? void 0 : data[index][1];
    }
    function listCacheHas(key) {
      return assocIndexOf(this.__data__, key) > -1;
    }
    function listCacheSet(key, value) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        data.push([key, value]);
      } else {
        data[index][1] = value;
      }
      return this;
    }
    ListCache.prototype.clear = listCacheClear;
    ListCache.prototype["delete"] = listCacheDelete;
    ListCache.prototype.get = listCacheGet;
    ListCache.prototype.has = listCacheHas;
    ListCache.prototype.set = listCacheSet;
    function MapCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function mapCacheClear() {
      this.__data__ = {
        "hash": new Hash(),
        "map": new (Map || ListCache)(),
        "string": new Hash()
      };
    }
    function mapCacheDelete(key) {
      return getMapData(this, key)["delete"](key);
    }
    function mapCacheGet(key) {
      return getMapData(this, key).get(key);
    }
    function mapCacheHas(key) {
      return getMapData(this, key).has(key);
    }
    function mapCacheSet(key, value) {
      getMapData(this, key).set(key, value);
      return this;
    }
    MapCache.prototype.clear = mapCacheClear;
    MapCache.prototype["delete"] = mapCacheDelete;
    MapCache.prototype.get = mapCacheGet;
    MapCache.prototype.has = mapCacheHas;
    MapCache.prototype.set = mapCacheSet;
    function SetCache(values) {
      var index = -1, length = values ? values.length : 0;
      this.__data__ = new MapCache();
      while (++index < length) {
        this.add(values[index]);
      }
    }
    function setCacheAdd(value) {
      this.__data__.set(value, HASH_UNDEFINED);
      return this;
    }
    function setCacheHas(value) {
      return this.__data__.has(value);
    }
    SetCache.prototype.add = SetCache.prototype.push = setCacheAdd;
    SetCache.prototype.has = setCacheHas;
    function Stack(entries) {
      this.__data__ = new ListCache(entries);
    }
    function stackClear() {
      this.__data__ = new ListCache();
    }
    function stackDelete(key) {
      return this.__data__["delete"](key);
    }
    function stackGet(key) {
      return this.__data__.get(key);
    }
    function stackHas(key) {
      return this.__data__.has(key);
    }
    function stackSet(key, value) {
      var cache = this.__data__;
      if (cache instanceof ListCache) {
        var pairs = cache.__data__;
        if (!Map || pairs.length < LARGE_ARRAY_SIZE - 1) {
          pairs.push([key, value]);
          return this;
        }
        cache = this.__data__ = new MapCache(pairs);
      }
      cache.set(key, value);
      return this;
    }
    Stack.prototype.clear = stackClear;
    Stack.prototype["delete"] = stackDelete;
    Stack.prototype.get = stackGet;
    Stack.prototype.has = stackHas;
    Stack.prototype.set = stackSet;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assocIndexOf(array, key) {
      var length = array.length;
      while (length--) {
        if (eq(array[length][0], key)) {
          return length;
        }
      }
      return -1;
    }
    var baseEach = createBaseEach(baseForOwn);
    function baseFilter(collection, predicate) {
      var result = [];
      baseEach(collection, function(value, index, collection2) {
        if (predicate(value, index, collection2)) {
          result.push(value);
        }
      });
      return result;
    }
    var baseFor = createBaseFor();
    function baseForOwn(object, iteratee) {
      return object && baseFor(object, iteratee, keys);
    }
    function baseGet(object, path) {
      path = isKey(path, object) ? [path] : castPath(path);
      var index = 0, length = path.length;
      while (object != null && index < length) {
        object = object[toKey(path[index++])];
      }
      return index && index == length ? object : void 0;
    }
    function baseGetTag(value) {
      return objectToString.call(value);
    }
    function baseHasIn(object, key) {
      return object != null && key in Object(object);
    }
    function baseIsEqual(value, other, customizer, bitmask, stack) {
      if (value === other) {
        return true;
      }
      if (value == null || other == null || !isObject(value) && !isObjectLike(other)) {
        return value !== value && other !== other;
      }
      return baseIsEqualDeep(value, other, baseIsEqual, customizer, bitmask, stack);
    }
    function baseIsEqualDeep(object, other, equalFunc, customizer, bitmask, stack) {
      var objIsArr = isArray(object), othIsArr = isArray(other), objTag = arrayTag, othTag = arrayTag;
      if (!objIsArr) {
        objTag = getTag(object);
        objTag = objTag == argsTag ? objectTag : objTag;
      }
      if (!othIsArr) {
        othTag = getTag(other);
        othTag = othTag == argsTag ? objectTag : othTag;
      }
      var objIsObj = objTag == objectTag && !isHostObject(object), othIsObj = othTag == objectTag && !isHostObject(other), isSameTag = objTag == othTag;
      if (isSameTag && !objIsObj) {
        stack || (stack = new Stack());
        return objIsArr || isTypedArray(object) ? equalArrays(object, other, equalFunc, customizer, bitmask, stack) : equalByTag(object, other, objTag, equalFunc, customizer, bitmask, stack);
      }
      if (!(bitmask & PARTIAL_COMPARE_FLAG)) {
        var objIsWrapped = objIsObj && hasOwnProperty2.call(object, "__wrapped__"), othIsWrapped = othIsObj && hasOwnProperty2.call(other, "__wrapped__");
        if (objIsWrapped || othIsWrapped) {
          var objUnwrapped = objIsWrapped ? object.value() : object, othUnwrapped = othIsWrapped ? other.value() : other;
          stack || (stack = new Stack());
          return equalFunc(objUnwrapped, othUnwrapped, customizer, bitmask, stack);
        }
      }
      if (!isSameTag) {
        return false;
      }
      stack || (stack = new Stack());
      return equalObjects(object, other, equalFunc, customizer, bitmask, stack);
    }
    function baseIsMatch(object, source, matchData, customizer) {
      var index = matchData.length, length = index, noCustomizer = !customizer;
      if (object == null) {
        return !length;
      }
      object = Object(object);
      while (index--) {
        var data = matchData[index];
        if (noCustomizer && data[2] ? data[1] !== object[data[0]] : !(data[0] in object)) {
          return false;
        }
      }
      while (++index < length) {
        data = matchData[index];
        var key = data[0], objValue = object[key], srcValue = data[1];
        if (noCustomizer && data[2]) {
          if (objValue === void 0 && !(key in object)) {
            return false;
          }
        } else {
          var stack = new Stack();
          if (customizer) {
            var result = customizer(objValue, srcValue, key, object, source, stack);
          }
          if (!(result === void 0 ? baseIsEqual(srcValue, objValue, customizer, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG, stack) : result)) {
            return false;
          }
        }
      }
      return true;
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) || isHostObject(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseIsTypedArray(value) {
      return isObjectLike(value) && isLength(value.length) && !!typedArrayTags[objectToString.call(value)];
    }
    function baseIteratee(value) {
      if (typeof value == "function") {
        return value;
      }
      if (value == null) {
        return identity;
      }
      if (typeof value == "object") {
        return isArray(value) ? baseMatchesProperty(value[0], value[1]) : baseMatches(value);
      }
      return property(value);
    }
    function baseKeys(object) {
      if (!isPrototype(object)) {
        return nativeKeys(object);
      }
      var result = [];
      for (var key in Object(object)) {
        if (hasOwnProperty2.call(object, key) && key != "constructor") {
          result.push(key);
        }
      }
      return result;
    }
    function baseMatches(source) {
      var matchData = getMatchData(source);
      if (matchData.length == 1 && matchData[0][2]) {
        return matchesStrictComparable(matchData[0][0], matchData[0][1]);
      }
      return function(object) {
        return object === source || baseIsMatch(object, source, matchData);
      };
    }
    function baseMatchesProperty(path, srcValue) {
      if (isKey(path) && isStrictComparable(srcValue)) {
        return matchesStrictComparable(toKey(path), srcValue);
      }
      return function(object) {
        var objValue = get(object, path);
        return objValue === void 0 && objValue === srcValue ? hasIn(object, path) : baseIsEqual(srcValue, objValue, void 0, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG);
      };
    }
    function basePropertyDeep(path) {
      return function(object) {
        return baseGet(object, path);
      };
    }
    function baseToString(value) {
      if (typeof value == "string") {
        return value;
      }
      if (isSymbol(value)) {
        return symbolToString ? symbolToString.call(value) : "";
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function castPath(value) {
      return isArray(value) ? value : stringToPath(value);
    }
    function createBaseEach(eachFunc, fromRight) {
      return function(collection, iteratee) {
        if (collection == null) {
          return collection;
        }
        if (!isArrayLike(collection)) {
          return eachFunc(collection, iteratee);
        }
        var length = collection.length, index = fromRight ? length : -1, iterable = Object(collection);
        while (fromRight ? index-- : ++index < length) {
          if (iteratee(iterable[index], index, iterable) === false) {
            break;
          }
        }
        return collection;
      };
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function equalArrays(array, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, arrLength = array.length, othLength = other.length;
      if (arrLength != othLength && !(isPartial && othLength > arrLength)) {
        return false;
      }
      var stacked = stack.get(array);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var index = -1, result = true, seen = bitmask & UNORDERED_COMPARE_FLAG ? new SetCache() : void 0;
      stack.set(array, other);
      stack.set(other, array);
      while (++index < arrLength) {
        var arrValue = array[index], othValue = other[index];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, arrValue, index, other, array, stack) : customizer(arrValue, othValue, index, array, other, stack);
        }
        if (compared !== void 0) {
          if (compared) {
            continue;
          }
          result = false;
          break;
        }
        if (seen) {
          if (!arraySome(other, function(othValue2, othIndex) {
            if (!seen.has(othIndex) && (arrValue === othValue2 || equalFunc(arrValue, othValue2, customizer, bitmask, stack))) {
              return seen.add(othIndex);
            }
          })) {
            result = false;
            break;
          }
        } else if (!(arrValue === othValue || equalFunc(arrValue, othValue, customizer, bitmask, stack))) {
          result = false;
          break;
        }
      }
      stack["delete"](array);
      stack["delete"](other);
      return result;
    }
    function equalByTag(object, other, tag, equalFunc, customizer, bitmask, stack) {
      switch (tag) {
        case dataViewTag:
          if (object.byteLength != other.byteLength || object.byteOffset != other.byteOffset) {
            return false;
          }
          object = object.buffer;
          other = other.buffer;
        case arrayBufferTag:
          if (object.byteLength != other.byteLength || !equalFunc(new Uint8Array2(object), new Uint8Array2(other))) {
            return false;
          }
          return true;
        case boolTag:
        case dateTag:
        case numberTag:
          return eq(+object, +other);
        case errorTag:
          return object.name == other.name && object.message == other.message;
        case regexpTag:
        case stringTag:
          return object == other + "";
        case mapTag:
          var convert = mapToArray;
        case setTag:
          var isPartial = bitmask & PARTIAL_COMPARE_FLAG;
          convert || (convert = setToArray);
          if (object.size != other.size && !isPartial) {
            return false;
          }
          var stacked = stack.get(object);
          if (stacked) {
            return stacked == other;
          }
          bitmask |= UNORDERED_COMPARE_FLAG;
          stack.set(object, other);
          var result = equalArrays(convert(object), convert(other), equalFunc, customizer, bitmask, stack);
          stack["delete"](object);
          return result;
        case symbolTag:
          if (symbolValueOf) {
            return symbolValueOf.call(object) == symbolValueOf.call(other);
          }
      }
      return false;
    }
    function equalObjects(object, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, objProps = keys(object), objLength = objProps.length, othProps = keys(other), othLength = othProps.length;
      if (objLength != othLength && !isPartial) {
        return false;
      }
      var index = objLength;
      while (index--) {
        var key = objProps[index];
        if (!(isPartial ? key in other : hasOwnProperty2.call(other, key))) {
          return false;
        }
      }
      var stacked = stack.get(object);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var result = true;
      stack.set(object, other);
      stack.set(other, object);
      var skipCtor = isPartial;
      while (++index < objLength) {
        key = objProps[index];
        var objValue = object[key], othValue = other[key];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, objValue, key, other, object, stack) : customizer(objValue, othValue, key, object, other, stack);
        }
        if (!(compared === void 0 ? objValue === othValue || equalFunc(objValue, othValue, customizer, bitmask, stack) : compared)) {
          result = false;
          break;
        }
        skipCtor || (skipCtor = key == "constructor");
      }
      if (result && !skipCtor) {
        var objCtor = object.constructor, othCtor = other.constructor;
        if (objCtor != othCtor && ("constructor" in object && "constructor" in other) && !(typeof objCtor == "function" && objCtor instanceof objCtor && typeof othCtor == "function" && othCtor instanceof othCtor)) {
          result = false;
        }
      }
      stack["delete"](object);
      stack["delete"](other);
      return result;
    }
    function getMapData(map, key) {
      var data = map.__data__;
      return isKeyable(key) ? data[typeof key == "string" ? "string" : "hash"] : data.map;
    }
    function getMatchData(object) {
      var result = keys(object), length = result.length;
      while (length--) {
        var key = result[length], value = object[key];
        result[length] = [key, value, isStrictComparable(value)];
      }
      return result;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    var getTag = baseGetTag;
    if (DataView && getTag(new DataView(new ArrayBuffer(1))) != dataViewTag || Map && getTag(new Map()) != mapTag || Promise2 && getTag(Promise2.resolve()) != promiseTag || Set && getTag(new Set()) != setTag || WeakMap && getTag(new WeakMap()) != weakMapTag) {
      getTag = function(value) {
        var result = objectToString.call(value), Ctor = result == objectTag ? value.constructor : void 0, ctorString = Ctor ? toSource(Ctor) : void 0;
        if (ctorString) {
          switch (ctorString) {
            case dataViewCtorString:
              return dataViewTag;
            case mapCtorString:
              return mapTag;
            case promiseCtorString:
              return promiseTag;
            case setCtorString:
              return setTag;
            case weakMapCtorString:
              return weakMapTag;
          }
        }
        return result;
      };
    }
    function hasPath(object, path, hasFunc) {
      path = isKey(path, object) ? [path] : castPath(path);
      var result, index = -1, length = path.length;
      while (++index < length) {
        var key = toKey(path[index]);
        if (!(result = object != null && hasFunc(object, key))) {
          break;
        }
        object = object[key];
      }
      if (result) {
        return result;
      }
      var length = object ? object.length : 0;
      return !!length && isLength(length) && isIndex(key, length) && (isArray(object) || isArguments(object));
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isKey(value, object) {
      if (isArray(value)) {
        return false;
      }
      var type = typeof value;
      if (type == "number" || type == "symbol" || type == "boolean" || value == null || isSymbol(value)) {
        return true;
      }
      return reIsPlainProp.test(value) || !reIsDeepProp.test(value) || object != null && value in Object(object);
    }
    function isKeyable(value) {
      var type = typeof value;
      return type == "string" || type == "number" || type == "symbol" || type == "boolean" ? value !== "__proto__" : value === null;
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function isStrictComparable(value) {
      return value === value && !isObject(value);
    }
    function matchesStrictComparable(key, srcValue) {
      return function(object) {
        if (object == null) {
          return false;
        }
        return object[key] === srcValue && (srcValue !== void 0 || key in Object(object));
      };
    }
    var stringToPath = memoize(function(string) {
      string = toString(string);
      var result = [];
      if (reLeadingDot.test(string)) {
        result.push("");
      }
      string.replace(rePropName, function(match, number, quote, string2) {
        result.push(quote ? string2.replace(reEscapeChar, "$1") : number || match);
      });
      return result;
    });
    function toKey(value) {
      if (typeof value == "string" || isSymbol(value)) {
        return value;
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function filter(collection, predicate) {
      var func = isArray(collection) ? arrayFilter : baseFilter;
      return func(collection, baseIteratee(predicate, 3));
    }
    function memoize(func, resolver) {
      if (typeof func != "function" || resolver && typeof resolver != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      var memoized = function() {
        var args = arguments, key = resolver ? resolver.apply(this, args) : args[0], cache = memoized.cache;
        if (cache.has(key)) {
          return cache.get(key);
        }
        var result = func.apply(this, args);
        memoized.cache = cache.set(key, result);
        return result;
      };
      memoized.cache = new (memoize.Cache || MapCache)();
      return memoized;
    }
    memoize.Cache = MapCache;
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    var isTypedArray = nodeIsTypedArray ? baseUnary(nodeIsTypedArray) : baseIsTypedArray;
    function toString(value) {
      return value == null ? "" : baseToString(value);
    }
    function get(object, path, defaultValue) {
      var result = object == null ? void 0 : baseGet(object, path);
      return result === void 0 ? defaultValue : result;
    }
    function hasIn(object, path) {
      return object != null && hasPath(object, path, baseHasIn);
    }
    function keys(object) {
      return isArrayLike(object) ? arrayLikeKeys(object) : baseKeys(object);
    }
    function identity(value) {
      return value;
    }
    function property(path) {
      return isKey(path) ? baseProperty(toKey(path)) : basePropertyDeep(path);
    }
    module2.exports = filter;
  }
});

// node_modules/lodash.reduce/index.js
var require_lodash9 = __commonJS({
  "node_modules/lodash.reduce/index.js"(exports2, module2) {
    init_process_shim();
    var LARGE_ARRAY_SIZE = 200;
    var FUNC_ERROR_TEXT = "Expected a function";
    var HASH_UNDEFINED = "__lodash_hash_undefined__";
    var UNORDERED_COMPARE_FLAG = 1;
    var PARTIAL_COMPARE_FLAG = 2;
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var arrayTag = "[object Array]";
    var boolTag = "[object Boolean]";
    var dateTag = "[object Date]";
    var errorTag = "[object Error]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var mapTag = "[object Map]";
    var numberTag = "[object Number]";
    var objectTag = "[object Object]";
    var promiseTag = "[object Promise]";
    var regexpTag = "[object RegExp]";
    var setTag = "[object Set]";
    var stringTag = "[object String]";
    var symbolTag = "[object Symbol]";
    var weakMapTag = "[object WeakMap]";
    var arrayBufferTag = "[object ArrayBuffer]";
    var dataViewTag = "[object DataView]";
    var float32Tag = "[object Float32Array]";
    var float64Tag = "[object Float64Array]";
    var int8Tag = "[object Int8Array]";
    var int16Tag = "[object Int16Array]";
    var int32Tag = "[object Int32Array]";
    var uint8Tag = "[object Uint8Array]";
    var uint8ClampedTag = "[object Uint8ClampedArray]";
    var uint16Tag = "[object Uint16Array]";
    var uint32Tag = "[object Uint32Array]";
    var reIsDeepProp = /\.|\[(?:[^[\]]*|(["'])(?:(?!\1)[^\\]|\\.)*?\1)\]/;
    var reIsPlainProp = /^\w*$/;
    var reLeadingDot = /^\./;
    var rePropName = /[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g;
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reEscapeChar = /\\(\\)?/g;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var typedArrayTags = {};
    typedArrayTags[float32Tag] = typedArrayTags[float64Tag] = typedArrayTags[int8Tag] = typedArrayTags[int16Tag] = typedArrayTags[int32Tag] = typedArrayTags[uint8Tag] = typedArrayTags[uint8ClampedTag] = typedArrayTags[uint16Tag] = typedArrayTags[uint32Tag] = true;
    typedArrayTags[argsTag] = typedArrayTags[arrayTag] = typedArrayTags[arrayBufferTag] = typedArrayTags[boolTag] = typedArrayTags[dataViewTag] = typedArrayTags[dateTag] = typedArrayTags[errorTag] = typedArrayTags[funcTag] = typedArrayTags[mapTag] = typedArrayTags[numberTag] = typedArrayTags[objectTag] = typedArrayTags[regexpTag] = typedArrayTags[setTag] = typedArrayTags[stringTag] = typedArrayTags[weakMapTag] = false;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    var freeExports = typeof exports2 == "object" && exports2 && !exports2.nodeType && exports2;
    var freeModule = freeExports && typeof module2 == "object" && module2 && !module2.nodeType && module2;
    var moduleExports = freeModule && freeModule.exports === freeExports;
    var freeProcess = moduleExports && freeGlobal.process;
    var nodeUtil = (function() {
      try {
        return freeProcess && freeProcess.binding("util");
      } catch (e) {
      }
    })();
    var nodeIsTypedArray = nodeUtil && nodeUtil.isTypedArray;
    function arrayReduce(array, iteratee, accumulator, initAccum) {
      var index = -1, length = array ? array.length : 0;
      if (initAccum && length) {
        accumulator = array[++index];
      }
      while (++index < length) {
        accumulator = iteratee(accumulator, array[index], index, array);
      }
      return accumulator;
    }
    function arraySome(array, predicate) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (predicate(array[index], index, array)) {
          return true;
        }
      }
      return false;
    }
    function baseProperty(key) {
      return function(object) {
        return object == null ? void 0 : object[key];
      };
    }
    function baseReduce(collection, iteratee, accumulator, initAccum, eachFunc) {
      eachFunc(collection, function(value, index, collection2) {
        accumulator = initAccum ? (initAccum = false, value) : iteratee(accumulator, value, index, collection2);
      });
      return accumulator;
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function baseUnary(func) {
      return function(value) {
        return func(value);
      };
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function isHostObject(value) {
      var result = false;
      if (value != null && typeof value.toString != "function") {
        try {
          result = !!(value + "");
        } catch (e) {
        }
      }
      return result;
    }
    function mapToArray(map) {
      var index = -1, result = Array(map.size);
      map.forEach(function(value, key) {
        result[++index] = [key, value];
      });
      return result;
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    function setToArray(set) {
      var index = -1, result = Array(set.size);
      set.forEach(function(value) {
        result[++index] = value;
      });
      return result;
    }
    var arrayProto = Array.prototype;
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var Symbol2 = root.Symbol;
    var Uint8Array2 = root.Uint8Array;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var splice = arrayProto.splice;
    var nativeKeys = overArg(Object.keys, Object);
    var DataView = getNative(root, "DataView");
    var Map = getNative(root, "Map");
    var Promise2 = getNative(root, "Promise");
    var Set = getNative(root, "Set");
    var WeakMap = getNative(root, "WeakMap");
    var nativeCreate = getNative(Object, "create");
    var dataViewCtorString = toSource(DataView);
    var mapCtorString = toSource(Map);
    var promiseCtorString = toSource(Promise2);
    var setCtorString = toSource(Set);
    var weakMapCtorString = toSource(WeakMap);
    var symbolProto = Symbol2 ? Symbol2.prototype : void 0;
    var symbolValueOf = symbolProto ? symbolProto.valueOf : void 0;
    var symbolToString = symbolProto ? symbolProto.toString : void 0;
    function Hash(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function hashClear() {
      this.__data__ = nativeCreate ? nativeCreate(null) : {};
    }
    function hashDelete(key) {
      return this.has(key) && delete this.__data__[key];
    }
    function hashGet(key) {
      var data = this.__data__;
      if (nativeCreate) {
        var result = data[key];
        return result === HASH_UNDEFINED ? void 0 : result;
      }
      return hasOwnProperty2.call(data, key) ? data[key] : void 0;
    }
    function hashHas(key) {
      var data = this.__data__;
      return nativeCreate ? data[key] !== void 0 : hasOwnProperty2.call(data, key);
    }
    function hashSet(key, value) {
      var data = this.__data__;
      data[key] = nativeCreate && value === void 0 ? HASH_UNDEFINED : value;
      return this;
    }
    Hash.prototype.clear = hashClear;
    Hash.prototype["delete"] = hashDelete;
    Hash.prototype.get = hashGet;
    Hash.prototype.has = hashHas;
    Hash.prototype.set = hashSet;
    function ListCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function listCacheClear() {
      this.__data__ = [];
    }
    function listCacheDelete(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        return false;
      }
      var lastIndex = data.length - 1;
      if (index == lastIndex) {
        data.pop();
      } else {
        splice.call(data, index, 1);
      }
      return true;
    }
    function listCacheGet(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      return index < 0 ? void 0 : data[index][1];
    }
    function listCacheHas(key) {
      return assocIndexOf(this.__data__, key) > -1;
    }
    function listCacheSet(key, value) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        data.push([key, value]);
      } else {
        data[index][1] = value;
      }
      return this;
    }
    ListCache.prototype.clear = listCacheClear;
    ListCache.prototype["delete"] = listCacheDelete;
    ListCache.prototype.get = listCacheGet;
    ListCache.prototype.has = listCacheHas;
    ListCache.prototype.set = listCacheSet;
    function MapCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function mapCacheClear() {
      this.__data__ = {
        "hash": new Hash(),
        "map": new (Map || ListCache)(),
        "string": new Hash()
      };
    }
    function mapCacheDelete(key) {
      return getMapData(this, key)["delete"](key);
    }
    function mapCacheGet(key) {
      return getMapData(this, key).get(key);
    }
    function mapCacheHas(key) {
      return getMapData(this, key).has(key);
    }
    function mapCacheSet(key, value) {
      getMapData(this, key).set(key, value);
      return this;
    }
    MapCache.prototype.clear = mapCacheClear;
    MapCache.prototype["delete"] = mapCacheDelete;
    MapCache.prototype.get = mapCacheGet;
    MapCache.prototype.has = mapCacheHas;
    MapCache.prototype.set = mapCacheSet;
    function SetCache(values) {
      var index = -1, length = values ? values.length : 0;
      this.__data__ = new MapCache();
      while (++index < length) {
        this.add(values[index]);
      }
    }
    function setCacheAdd(value) {
      this.__data__.set(value, HASH_UNDEFINED);
      return this;
    }
    function setCacheHas(value) {
      return this.__data__.has(value);
    }
    SetCache.prototype.add = SetCache.prototype.push = setCacheAdd;
    SetCache.prototype.has = setCacheHas;
    function Stack(entries) {
      this.__data__ = new ListCache(entries);
    }
    function stackClear() {
      this.__data__ = new ListCache();
    }
    function stackDelete(key) {
      return this.__data__["delete"](key);
    }
    function stackGet(key) {
      return this.__data__.get(key);
    }
    function stackHas(key) {
      return this.__data__.has(key);
    }
    function stackSet(key, value) {
      var cache = this.__data__;
      if (cache instanceof ListCache) {
        var pairs = cache.__data__;
        if (!Map || pairs.length < LARGE_ARRAY_SIZE - 1) {
          pairs.push([key, value]);
          return this;
        }
        cache = this.__data__ = new MapCache(pairs);
      }
      cache.set(key, value);
      return this;
    }
    Stack.prototype.clear = stackClear;
    Stack.prototype["delete"] = stackDelete;
    Stack.prototype.get = stackGet;
    Stack.prototype.has = stackHas;
    Stack.prototype.set = stackSet;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assocIndexOf(array, key) {
      var length = array.length;
      while (length--) {
        if (eq(array[length][0], key)) {
          return length;
        }
      }
      return -1;
    }
    var baseEach = createBaseEach(baseForOwn);
    var baseFor = createBaseFor();
    function baseForOwn(object, iteratee) {
      return object && baseFor(object, iteratee, keys);
    }
    function baseGet(object, path) {
      path = isKey(path, object) ? [path] : castPath(path);
      var index = 0, length = path.length;
      while (object != null && index < length) {
        object = object[toKey(path[index++])];
      }
      return index && index == length ? object : void 0;
    }
    function baseGetTag(value) {
      return objectToString.call(value);
    }
    function baseHasIn(object, key) {
      return object != null && key in Object(object);
    }
    function baseIsEqual(value, other, customizer, bitmask, stack) {
      if (value === other) {
        return true;
      }
      if (value == null || other == null || !isObject(value) && !isObjectLike(other)) {
        return value !== value && other !== other;
      }
      return baseIsEqualDeep(value, other, baseIsEqual, customizer, bitmask, stack);
    }
    function baseIsEqualDeep(object, other, equalFunc, customizer, bitmask, stack) {
      var objIsArr = isArray(object), othIsArr = isArray(other), objTag = arrayTag, othTag = arrayTag;
      if (!objIsArr) {
        objTag = getTag(object);
        objTag = objTag == argsTag ? objectTag : objTag;
      }
      if (!othIsArr) {
        othTag = getTag(other);
        othTag = othTag == argsTag ? objectTag : othTag;
      }
      var objIsObj = objTag == objectTag && !isHostObject(object), othIsObj = othTag == objectTag && !isHostObject(other), isSameTag = objTag == othTag;
      if (isSameTag && !objIsObj) {
        stack || (stack = new Stack());
        return objIsArr || isTypedArray(object) ? equalArrays(object, other, equalFunc, customizer, bitmask, stack) : equalByTag(object, other, objTag, equalFunc, customizer, bitmask, stack);
      }
      if (!(bitmask & PARTIAL_COMPARE_FLAG)) {
        var objIsWrapped = objIsObj && hasOwnProperty2.call(object, "__wrapped__"), othIsWrapped = othIsObj && hasOwnProperty2.call(other, "__wrapped__");
        if (objIsWrapped || othIsWrapped) {
          var objUnwrapped = objIsWrapped ? object.value() : object, othUnwrapped = othIsWrapped ? other.value() : other;
          stack || (stack = new Stack());
          return equalFunc(objUnwrapped, othUnwrapped, customizer, bitmask, stack);
        }
      }
      if (!isSameTag) {
        return false;
      }
      stack || (stack = new Stack());
      return equalObjects(object, other, equalFunc, customizer, bitmask, stack);
    }
    function baseIsMatch(object, source, matchData, customizer) {
      var index = matchData.length, length = index, noCustomizer = !customizer;
      if (object == null) {
        return !length;
      }
      object = Object(object);
      while (index--) {
        var data = matchData[index];
        if (noCustomizer && data[2] ? data[1] !== object[data[0]] : !(data[0] in object)) {
          return false;
        }
      }
      while (++index < length) {
        data = matchData[index];
        var key = data[0], objValue = object[key], srcValue = data[1];
        if (noCustomizer && data[2]) {
          if (objValue === void 0 && !(key in object)) {
            return false;
          }
        } else {
          var stack = new Stack();
          if (customizer) {
            var result = customizer(objValue, srcValue, key, object, source, stack);
          }
          if (!(result === void 0 ? baseIsEqual(srcValue, objValue, customizer, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG, stack) : result)) {
            return false;
          }
        }
      }
      return true;
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) || isHostObject(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseIsTypedArray(value) {
      return isObjectLike(value) && isLength(value.length) && !!typedArrayTags[objectToString.call(value)];
    }
    function baseIteratee(value) {
      if (typeof value == "function") {
        return value;
      }
      if (value == null) {
        return identity;
      }
      if (typeof value == "object") {
        return isArray(value) ? baseMatchesProperty(value[0], value[1]) : baseMatches(value);
      }
      return property(value);
    }
    function baseKeys(object) {
      if (!isPrototype(object)) {
        return nativeKeys(object);
      }
      var result = [];
      for (var key in Object(object)) {
        if (hasOwnProperty2.call(object, key) && key != "constructor") {
          result.push(key);
        }
      }
      return result;
    }
    function baseMatches(source) {
      var matchData = getMatchData(source);
      if (matchData.length == 1 && matchData[0][2]) {
        return matchesStrictComparable(matchData[0][0], matchData[0][1]);
      }
      return function(object) {
        return object === source || baseIsMatch(object, source, matchData);
      };
    }
    function baseMatchesProperty(path, srcValue) {
      if (isKey(path) && isStrictComparable(srcValue)) {
        return matchesStrictComparable(toKey(path), srcValue);
      }
      return function(object) {
        var objValue = get(object, path);
        return objValue === void 0 && objValue === srcValue ? hasIn(object, path) : baseIsEqual(srcValue, objValue, void 0, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG);
      };
    }
    function basePropertyDeep(path) {
      return function(object) {
        return baseGet(object, path);
      };
    }
    function baseToString(value) {
      if (typeof value == "string") {
        return value;
      }
      if (isSymbol(value)) {
        return symbolToString ? symbolToString.call(value) : "";
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function castPath(value) {
      return isArray(value) ? value : stringToPath(value);
    }
    function createBaseEach(eachFunc, fromRight) {
      return function(collection, iteratee) {
        if (collection == null) {
          return collection;
        }
        if (!isArrayLike(collection)) {
          return eachFunc(collection, iteratee);
        }
        var length = collection.length, index = fromRight ? length : -1, iterable = Object(collection);
        while (fromRight ? index-- : ++index < length) {
          if (iteratee(iterable[index], index, iterable) === false) {
            break;
          }
        }
        return collection;
      };
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function equalArrays(array, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, arrLength = array.length, othLength = other.length;
      if (arrLength != othLength && !(isPartial && othLength > arrLength)) {
        return false;
      }
      var stacked = stack.get(array);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var index = -1, result = true, seen = bitmask & UNORDERED_COMPARE_FLAG ? new SetCache() : void 0;
      stack.set(array, other);
      stack.set(other, array);
      while (++index < arrLength) {
        var arrValue = array[index], othValue = other[index];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, arrValue, index, other, array, stack) : customizer(arrValue, othValue, index, array, other, stack);
        }
        if (compared !== void 0) {
          if (compared) {
            continue;
          }
          result = false;
          break;
        }
        if (seen) {
          if (!arraySome(other, function(othValue2, othIndex) {
            if (!seen.has(othIndex) && (arrValue === othValue2 || equalFunc(arrValue, othValue2, customizer, bitmask, stack))) {
              return seen.add(othIndex);
            }
          })) {
            result = false;
            break;
          }
        } else if (!(arrValue === othValue || equalFunc(arrValue, othValue, customizer, bitmask, stack))) {
          result = false;
          break;
        }
      }
      stack["delete"](array);
      stack["delete"](other);
      return result;
    }
    function equalByTag(object, other, tag, equalFunc, customizer, bitmask, stack) {
      switch (tag) {
        case dataViewTag:
          if (object.byteLength != other.byteLength || object.byteOffset != other.byteOffset) {
            return false;
          }
          object = object.buffer;
          other = other.buffer;
        case arrayBufferTag:
          if (object.byteLength != other.byteLength || !equalFunc(new Uint8Array2(object), new Uint8Array2(other))) {
            return false;
          }
          return true;
        case boolTag:
        case dateTag:
        case numberTag:
          return eq(+object, +other);
        case errorTag:
          return object.name == other.name && object.message == other.message;
        case regexpTag:
        case stringTag:
          return object == other + "";
        case mapTag:
          var convert = mapToArray;
        case setTag:
          var isPartial = bitmask & PARTIAL_COMPARE_FLAG;
          convert || (convert = setToArray);
          if (object.size != other.size && !isPartial) {
            return false;
          }
          var stacked = stack.get(object);
          if (stacked) {
            return stacked == other;
          }
          bitmask |= UNORDERED_COMPARE_FLAG;
          stack.set(object, other);
          var result = equalArrays(convert(object), convert(other), equalFunc, customizer, bitmask, stack);
          stack["delete"](object);
          return result;
        case symbolTag:
          if (symbolValueOf) {
            return symbolValueOf.call(object) == symbolValueOf.call(other);
          }
      }
      return false;
    }
    function equalObjects(object, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, objProps = keys(object), objLength = objProps.length, othProps = keys(other), othLength = othProps.length;
      if (objLength != othLength && !isPartial) {
        return false;
      }
      var index = objLength;
      while (index--) {
        var key = objProps[index];
        if (!(isPartial ? key in other : hasOwnProperty2.call(other, key))) {
          return false;
        }
      }
      var stacked = stack.get(object);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var result = true;
      stack.set(object, other);
      stack.set(other, object);
      var skipCtor = isPartial;
      while (++index < objLength) {
        key = objProps[index];
        var objValue = object[key], othValue = other[key];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, objValue, key, other, object, stack) : customizer(objValue, othValue, key, object, other, stack);
        }
        if (!(compared === void 0 ? objValue === othValue || equalFunc(objValue, othValue, customizer, bitmask, stack) : compared)) {
          result = false;
          break;
        }
        skipCtor || (skipCtor = key == "constructor");
      }
      if (result && !skipCtor) {
        var objCtor = object.constructor, othCtor = other.constructor;
        if (objCtor != othCtor && ("constructor" in object && "constructor" in other) && !(typeof objCtor == "function" && objCtor instanceof objCtor && typeof othCtor == "function" && othCtor instanceof othCtor)) {
          result = false;
        }
      }
      stack["delete"](object);
      stack["delete"](other);
      return result;
    }
    function getMapData(map, key) {
      var data = map.__data__;
      return isKeyable(key) ? data[typeof key == "string" ? "string" : "hash"] : data.map;
    }
    function getMatchData(object) {
      var result = keys(object), length = result.length;
      while (length--) {
        var key = result[length], value = object[key];
        result[length] = [key, value, isStrictComparable(value)];
      }
      return result;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    var getTag = baseGetTag;
    if (DataView && getTag(new DataView(new ArrayBuffer(1))) != dataViewTag || Map && getTag(new Map()) != mapTag || Promise2 && getTag(Promise2.resolve()) != promiseTag || Set && getTag(new Set()) != setTag || WeakMap && getTag(new WeakMap()) != weakMapTag) {
      getTag = function(value) {
        var result = objectToString.call(value), Ctor = result == objectTag ? value.constructor : void 0, ctorString = Ctor ? toSource(Ctor) : void 0;
        if (ctorString) {
          switch (ctorString) {
            case dataViewCtorString:
              return dataViewTag;
            case mapCtorString:
              return mapTag;
            case promiseCtorString:
              return promiseTag;
            case setCtorString:
              return setTag;
            case weakMapCtorString:
              return weakMapTag;
          }
        }
        return result;
      };
    }
    function hasPath(object, path, hasFunc) {
      path = isKey(path, object) ? [path] : castPath(path);
      var result, index = -1, length = path.length;
      while (++index < length) {
        var key = toKey(path[index]);
        if (!(result = object != null && hasFunc(object, key))) {
          break;
        }
        object = object[key];
      }
      if (result) {
        return result;
      }
      var length = object ? object.length : 0;
      return !!length && isLength(length) && isIndex(key, length) && (isArray(object) || isArguments(object));
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isKey(value, object) {
      if (isArray(value)) {
        return false;
      }
      var type = typeof value;
      if (type == "number" || type == "symbol" || type == "boolean" || value == null || isSymbol(value)) {
        return true;
      }
      return reIsPlainProp.test(value) || !reIsDeepProp.test(value) || object != null && value in Object(object);
    }
    function isKeyable(value) {
      var type = typeof value;
      return type == "string" || type == "number" || type == "symbol" || type == "boolean" ? value !== "__proto__" : value === null;
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function isStrictComparable(value) {
      return value === value && !isObject(value);
    }
    function matchesStrictComparable(key, srcValue) {
      return function(object) {
        if (object == null) {
          return false;
        }
        return object[key] === srcValue && (srcValue !== void 0 || key in Object(object));
      };
    }
    var stringToPath = memoize(function(string) {
      string = toString(string);
      var result = [];
      if (reLeadingDot.test(string)) {
        result.push("");
      }
      string.replace(rePropName, function(match, number, quote, string2) {
        result.push(quote ? string2.replace(reEscapeChar, "$1") : number || match);
      });
      return result;
    });
    function toKey(value) {
      if (typeof value == "string" || isSymbol(value)) {
        return value;
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function reduce(collection, iteratee, accumulator) {
      var func = isArray(collection) ? arrayReduce : baseReduce, initAccum = arguments.length < 3;
      return func(collection, baseIteratee(iteratee, 4), accumulator, initAccum, baseEach);
    }
    function memoize(func, resolver) {
      if (typeof func != "function" || resolver && typeof resolver != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      var memoized = function() {
        var args = arguments, key = resolver ? resolver.apply(this, args) : args[0], cache = memoized.cache;
        if (cache.has(key)) {
          return cache.get(key);
        }
        var result = func.apply(this, args);
        memoized.cache = cache.set(key, result);
        return result;
      };
      memoized.cache = new (memoize.Cache || MapCache)();
      return memoized;
    }
    memoize.Cache = MapCache;
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    var isTypedArray = nodeIsTypedArray ? baseUnary(nodeIsTypedArray) : baseIsTypedArray;
    function toString(value) {
      return value == null ? "" : baseToString(value);
    }
    function get(object, path, defaultValue) {
      var result = object == null ? void 0 : baseGet(object, path);
      return result === void 0 ? defaultValue : result;
    }
    function hasIn(object, path) {
      return object != null && hasPath(object, path, baseHasIn);
    }
    function keys(object) {
      return isArrayLike(object) ? arrayLikeKeys(object) : baseKeys(object);
    }
    function identity(value) {
      return value;
    }
    function property(path) {
      return isKey(path) ? baseProperty(toKey(path)) : basePropertyDeep(path);
    }
    module2.exports = reduce;
  }
});

// node_modules/cheerio/lib/api/traversing.js
var require_traversing = __commonJS({
  "node_modules/cheerio/lib/api/traversing.js"(exports2) {
    init_process_shim();
    var select = require_css_select();
    var utils = require_utils();
    var domEach = utils.domEach;
    var uniqueSort = require_lib().DomUtils.uniqueSort;
    var isTag = utils.isTag;
    var _ = {
      bind: require_lodash2(),
      forEach: require_lodash3(),
      reject: require_lodash7(),
      filter: require_lodash8(),
      reduce: require_lodash9()
    };
    exports2.find = function(selectorOrHaystack) {
      var elems = _.reduce(this, function(memo, elem) {
        return memo.concat(_.filter(elem.children, isTag));
      }, []);
      var contains = this.constructor.contains;
      var haystack;
      if (selectorOrHaystack && typeof selectorOrHaystack !== "string") {
        if (selectorOrHaystack.cheerio) {
          haystack = selectorOrHaystack.get();
        } else {
          haystack = [selectorOrHaystack];
        }
        return this._make(haystack.filter(function(elem) {
          var idx, len;
          for (idx = 0, len = this.length; idx < len; ++idx) {
            if (contains(this[idx], elem)) {
              return true;
            }
          }
        }, this));
      }
      var options = { __proto__: this.options, context: this.toArray() };
      return this._make(select(selectorOrHaystack, elems, options));
    };
    exports2.parent = function(selector) {
      var set = [];
      domEach(this, function(idx, elem) {
        var parentElem = elem.parent;
        if (parentElem && set.indexOf(parentElem) < 0) {
          set.push(parentElem);
        }
      });
      if (arguments.length) {
        set = exports2.filter.call(set, selector, this);
      }
      return this._make(set);
    };
    exports2.parents = function(selector) {
      var parentNodes = [];
      this.get().reverse().forEach(function(elem) {
        traverseParents(this, elem.parent, selector, Infinity).forEach(
          function(node) {
            if (parentNodes.indexOf(node) === -1) {
              parentNodes.push(node);
            }
          }
        );
      }, this);
      return this._make(parentNodes);
    };
    exports2.parentsUntil = function(selector, filter) {
      var parentNodes = [], untilNode, untilNodes;
      if (typeof selector === "string") {
        untilNode = select(selector, this.parents().toArray(), this.options)[0];
      } else if (selector && selector.cheerio) {
        untilNodes = selector.toArray();
      } else if (selector) {
        untilNode = selector;
      }
      this.toArray().reverse().forEach(function(elem) {
        while (elem = elem.parent) {
          if (untilNode && elem !== untilNode || untilNodes && untilNodes.indexOf(elem) === -1 || !untilNode && !untilNodes) {
            if (isTag(elem) && parentNodes.indexOf(elem) === -1) {
              parentNodes.push(elem);
            }
          } else {
            break;
          }
        }
      }, this);
      return this._make(filter ? select(filter, parentNodes, this.options) : parentNodes);
    };
    exports2.closest = function(selector) {
      var set = [];
      if (!selector) {
        return this._make(set);
      }
      domEach(this, function(idx, elem) {
        var closestElem = traverseParents(this, elem, selector, 1)[0];
        if (closestElem && set.indexOf(closestElem) < 0) {
          set.push(closestElem);
        }
      }.bind(this));
      return this._make(set);
    };
    exports2.next = function(selector) {
      if (!this[0]) {
        return this;
      }
      var elems = [];
      _.forEach(this, function(elem) {
        while (elem = elem.next) {
          if (isTag(elem)) {
            elems.push(elem);
            return;
          }
        }
      });
      return selector ? exports2.filter.call(elems, selector, this) : this._make(elems);
    };
    exports2.nextAll = function(selector) {
      if (!this[0]) {
        return this;
      }
      var elems = [];
      _.forEach(this, function(elem) {
        while (elem = elem.next) {
          if (isTag(elem) && elems.indexOf(elem) === -1) {
            elems.push(elem);
          }
        }
      });
      return selector ? exports2.filter.call(elems, selector, this) : this._make(elems);
    };
    exports2.nextUntil = function(selector, filterSelector) {
      if (!this[0]) {
        return this;
      }
      var elems = [], untilNode, untilNodes;
      if (typeof selector === "string") {
        untilNode = select(selector, this.nextAll().get(), this.options)[0];
      } else if (selector && selector.cheerio) {
        untilNodes = selector.get();
      } else if (selector) {
        untilNode = selector;
      }
      _.forEach(this, function(elem) {
        while (elem = elem.next) {
          if (untilNode && elem !== untilNode || untilNodes && untilNodes.indexOf(elem) === -1 || !untilNode && !untilNodes) {
            if (isTag(elem) && elems.indexOf(elem) === -1) {
              elems.push(elem);
            }
          } else {
            break;
          }
        }
      });
      return filterSelector ? exports2.filter.call(elems, filterSelector, this) : this._make(elems);
    };
    exports2.prev = function(selector) {
      if (!this[0]) {
        return this;
      }
      var elems = [];
      _.forEach(this, function(elem) {
        while (elem = elem.prev) {
          if (isTag(elem)) {
            elems.push(elem);
            return;
          }
        }
      });
      return selector ? exports2.filter.call(elems, selector, this) : this._make(elems);
    };
    exports2.prevAll = function(selector) {
      if (!this[0]) {
        return this;
      }
      var elems = [];
      _.forEach(this, function(elem) {
        while (elem = elem.prev) {
          if (isTag(elem) && elems.indexOf(elem) === -1) {
            elems.push(elem);
          }
        }
      });
      return selector ? exports2.filter.call(elems, selector, this) : this._make(elems);
    };
    exports2.prevUntil = function(selector, filterSelector) {
      if (!this[0]) {
        return this;
      }
      var elems = [], untilNode, untilNodes;
      if (typeof selector === "string") {
        untilNode = select(selector, this.prevAll().get(), this.options)[0];
      } else if (selector && selector.cheerio) {
        untilNodes = selector.get();
      } else if (selector) {
        untilNode = selector;
      }
      _.forEach(this, function(elem) {
        while (elem = elem.prev) {
          if (untilNode && elem !== untilNode || untilNodes && untilNodes.indexOf(elem) === -1 || !untilNode && !untilNodes) {
            if (isTag(elem) && elems.indexOf(elem) === -1) {
              elems.push(elem);
            }
          } else {
            break;
          }
        }
      });
      return filterSelector ? exports2.filter.call(elems, filterSelector, this) : this._make(elems);
    };
    exports2.siblings = function(selector) {
      var parent = this.parent();
      var elems = _.filter(
        parent ? parent.children() : this.siblingsAndMe(),
        _.bind(function(elem) {
          return isTag(elem) && !this.is(elem);
        }, this)
      );
      if (selector !== void 0) {
        return exports2.filter.call(elems, selector, this);
      } else {
        return this._make(elems);
      }
    };
    exports2.children = function(selector) {
      var elems = _.reduce(this, function(memo, elem) {
        return memo.concat(_.filter(elem.children, isTag));
      }, []);
      if (selector === void 0) return this._make(elems);
      return exports2.filter.call(elems, selector, this);
    };
    exports2.contents = function() {
      return this._make(_.reduce(this, function(all, elem) {
        all.push.apply(all, elem.children);
        return all;
      }, []));
    };
    exports2.each = function(fn) {
      var i = 0, len = this.length;
      while (i < len && fn.call(this[i], i, this[i]) !== false) ++i;
      return this;
    };
    exports2.map = function(fn) {
      return this._make(_.reduce(this, function(memo, el, i) {
        var val = fn.call(el, i, el);
        return val == null ? memo : memo.concat(val);
      }, []));
    };
    var makeFilterMethod = function(filterFn) {
      return function(match, container) {
        var testFn;
        container = container || this;
        if (typeof match === "string") {
          testFn = select.compile(match, container.options);
        } else if (typeof match === "function") {
          testFn = function(el, i) {
            return match.call(el, i, el);
          };
        } else if (match.cheerio) {
          testFn = match.is.bind(match);
        } else {
          testFn = function(el) {
            return match === el;
          };
        }
        return container._make(filterFn(this, testFn));
      };
    };
    exports2.filter = makeFilterMethod(_.filter);
    exports2.not = makeFilterMethod(_.reject);
    exports2.has = function(selectorOrHaystack) {
      var that = this;
      return exports2.filter.call(this, function() {
        return that._make(this).find(selectorOrHaystack).length > 0;
      });
    };
    exports2.first = function() {
      return this.length > 1 ? this._make(this[0]) : this;
    };
    exports2.last = function() {
      return this.length > 1 ? this._make(this[this.length - 1]) : this;
    };
    exports2.eq = function(i) {
      i = +i;
      if (i === 0 && this.length <= 1) return this;
      if (i < 0) i = this.length + i;
      return this[i] ? this._make(this[i]) : this._make([]);
    };
    exports2.get = function(i) {
      if (i == null) {
        return Array.prototype.slice.call(this);
      } else {
        return this[i < 0 ? this.length + i : i];
      }
    };
    exports2.index = function(selectorOrNeedle) {
      var $haystack, needle;
      if (arguments.length === 0) {
        $haystack = this.parent().children();
        needle = this[0];
      } else if (typeof selectorOrNeedle === "string") {
        $haystack = this._make(selectorOrNeedle);
        needle = this[0];
      } else {
        $haystack = this;
        needle = selectorOrNeedle.cheerio ? selectorOrNeedle[0] : selectorOrNeedle;
      }
      return $haystack.get().indexOf(needle);
    };
    exports2.slice = function() {
      return this._make([].slice.apply(this, arguments));
    };
    function traverseParents(self2, elem, selector, limit) {
      var elems = [];
      while (elem && elems.length < limit) {
        if (!selector || exports2.filter.call([elem], selector, self2).length) {
          elems.push(elem);
        }
        elem = elem.parent;
      }
      return elems;
    }
    exports2.end = function() {
      return this.prevObject || this._make([]);
    };
    exports2.add = function(other, context) {
      var selection = this._make(other, context);
      var contents = uniqueSort(selection.get().concat(this.get()));
      for (var i = 0; i < contents.length; ++i) {
        selection[i] = contents[i];
      }
      selection.length = contents.length;
      return selection;
    };
    exports2.addBack = function(selector) {
      return this.add(
        arguments.length ? this.prevObject.filter(selector) : this.prevObject
      );
    };
  }
});

// node_modules/lodash.flatten/index.js
var require_lodash10 = __commonJS({
  "node_modules/lodash.flatten/index.js"(exports2, module2) {
    init_process_shim();
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    function arrayPush(array, values) {
      var index = -1, length = values.length, offset = array.length;
      while (++index < length) {
        array[offset + index] = values[index];
      }
      return array;
    }
    var objectProto = Object.prototype;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var Symbol2 = root.Symbol;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var spreadableSymbol = Symbol2 ? Symbol2.isConcatSpreadable : void 0;
    function baseFlatten(array, depth, predicate, isStrict, result) {
      var index = -1, length = array.length;
      predicate || (predicate = isFlattenable);
      result || (result = []);
      while (++index < length) {
        var value = array[index];
        if (depth > 0 && predicate(value)) {
          if (depth > 1) {
            baseFlatten(value, depth - 1, predicate, isStrict, result);
          } else {
            arrayPush(result, value);
          }
        } else if (!isStrict) {
          result[result.length] = value;
        }
      }
      return result;
    }
    function isFlattenable(value) {
      return isArray(value) || isArguments(value) || !!(spreadableSymbol && value && value[spreadableSymbol]);
    }
    function flatten(array) {
      var length = array ? array.length : 0;
      return length ? baseFlatten(array, 1) : [];
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    module2.exports = flatten;
  }
});

// node_modules/cheerio/lib/api/manipulation.js
var require_manipulation2 = __commonJS({
  "node_modules/cheerio/lib/api/manipulation.js"(exports2) {
    init_process_shim();
    var parse = require_parse();
    var $ = require_static();
    var updateDOM = parse.update;
    var evaluate = parse.evaluate;
    var utils = require_utils();
    var domEach = utils.domEach;
    var cloneDom = utils.cloneDom;
    var isHtml = utils.isHtml;
    var slice = Array.prototype.slice;
    var _ = {
      flatten: require_lodash10(),
      bind: require_lodash2(),
      forEach: require_lodash3()
    };
    exports2._makeDomArray = function makeDomArray(elem, clone) {
      if (elem == null) {
        return [];
      } else if (elem.cheerio) {
        return clone ? cloneDom(elem.get(), elem.options) : elem.get();
      } else if (Array.isArray(elem)) {
        return _.flatten(elem.map(function(el) {
          return this._makeDomArray(el, clone);
        }, this));
      } else if (typeof elem === "string") {
        return evaluate(elem, this.options);
      } else {
        return clone ? cloneDom([elem]) : [elem];
      }
    };
    var _insert = function(concatenator) {
      return function() {
        var elems = slice.call(arguments), lastIdx = this.length - 1;
        return domEach(this, function(i, el) {
          var dom, domSrc;
          if (typeof elems[0] === "function") {
            domSrc = elems[0].call(el, i, $.html(el.children));
          } else {
            domSrc = elems;
          }
          dom = this._makeDomArray(domSrc, i < lastIdx);
          concatenator(dom, el.children, el);
        });
      };
    };
    var uniqueSplice = function(array, spliceIdx, spliceCount, newElems, parent) {
      var spliceArgs = [spliceIdx, spliceCount].concat(newElems), prev = array[spliceIdx - 1] || null, next = array[spliceIdx] || null;
      var idx, len, prevIdx, node, oldParent;
      for (idx = 0, len = newElems.length; idx < len; ++idx) {
        node = newElems[idx];
        oldParent = node.parent || node.root;
        prevIdx = oldParent && oldParent.children.indexOf(newElems[idx]);
        if (oldParent && prevIdx > -1) {
          oldParent.children.splice(prevIdx, 1);
          if (parent === oldParent && spliceIdx > prevIdx) {
            spliceArgs[0]--;
          }
        }
        node.root = null;
        node.parent = parent;
        if (node.prev) {
          node.prev.next = node.next || null;
        }
        if (node.next) {
          node.next.prev = node.prev || null;
        }
        node.prev = newElems[idx - 1] || prev;
        node.next = newElems[idx + 1] || next;
      }
      if (prev) {
        prev.next = newElems[0];
      }
      if (next) {
        next.prev = newElems[newElems.length - 1];
      }
      return array.splice.apply(array, spliceArgs);
    };
    exports2.appendTo = function(target) {
      if (!target.cheerio) {
        target = this.constructor.call(this.constructor, target, null, this._originalRoot);
      }
      target.append(this);
      return this;
    };
    exports2.prependTo = function(target) {
      if (!target.cheerio) {
        target = this.constructor.call(this.constructor, target, null, this._originalRoot);
      }
      target.prepend(this);
      return this;
    };
    exports2.append = _insert(function(dom, children, parent) {
      uniqueSplice(children, children.length, 0, dom, parent);
    });
    exports2.prepend = _insert(function(dom, children, parent) {
      uniqueSplice(children, 0, 0, dom, parent);
    });
    exports2.wrap = function(wrapper) {
      var wrapperFn = typeof wrapper === "function" && wrapper, lastIdx = this.length - 1;
      _.forEach(this, _.bind(function(el, i) {
        var parent = el.parent || el.root, siblings = parent.children, dom, index;
        if (!parent) {
          return;
        }
        if (wrapperFn) {
          wrapper = wrapperFn.call(el, i);
        }
        if (typeof wrapper === "string" && !isHtml(wrapper)) {
          wrapper = this.parents().last().find(wrapper).clone();
        }
        dom = this._makeDomArray(wrapper, i < lastIdx).slice(0, 1);
        index = siblings.indexOf(el);
        updateDOM([el], dom[0]);
        uniqueSplice(siblings, index, 0, dom, parent);
      }, this));
      return this;
    };
    exports2.after = function() {
      var elems = slice.call(arguments), lastIdx = this.length - 1;
      domEach(this, function(i, el) {
        var parent = el.parent || el.root;
        if (!parent) {
          return;
        }
        var siblings = parent.children, index = siblings.indexOf(el), domSrc, dom;
        if (index < 0) return;
        if (typeof elems[0] === "function") {
          domSrc = elems[0].call(el, i, $.html(el.children));
        } else {
          domSrc = elems;
        }
        dom = this._makeDomArray(domSrc, i < lastIdx);
        uniqueSplice(siblings, index + 1, 0, dom, parent);
      });
      return this;
    };
    exports2.insertAfter = function(target) {
      var clones = [], self2 = this;
      if (typeof target === "string") {
        target = this.constructor.call(this.constructor, target, null, this._originalRoot);
      }
      target = this._makeDomArray(target);
      self2.remove();
      domEach(target, function(i, el) {
        var clonedSelf = self2._makeDomArray(self2.clone());
        var parent = el.parent || el.root;
        if (!parent) {
          return;
        }
        var siblings = parent.children, index = siblings.indexOf(el);
        if (index < 0) return;
        uniqueSplice(siblings, index + 1, 0, clonedSelf, parent);
        clones.push(clonedSelf);
      });
      return this.constructor.call(this.constructor, this._makeDomArray(clones));
    };
    exports2.before = function() {
      var elems = slice.call(arguments), lastIdx = this.length - 1;
      domEach(this, function(i, el) {
        var parent = el.parent || el.root;
        if (!parent) {
          return;
        }
        var siblings = parent.children, index = siblings.indexOf(el), domSrc, dom;
        if (index < 0) return;
        if (typeof elems[0] === "function") {
          domSrc = elems[0].call(el, i, $.html(el.children));
        } else {
          domSrc = elems;
        }
        dom = this._makeDomArray(domSrc, i < lastIdx);
        uniqueSplice(siblings, index, 0, dom, parent);
      });
      return this;
    };
    exports2.insertBefore = function(target) {
      var clones = [], self2 = this;
      if (typeof target === "string") {
        target = this.constructor.call(this.constructor, target, null, this._originalRoot);
      }
      target = this._makeDomArray(target);
      self2.remove();
      domEach(target, function(i, el) {
        var clonedSelf = self2._makeDomArray(self2.clone());
        var parent = el.parent || el.root;
        if (!parent) {
          return;
        }
        var siblings = parent.children, index = siblings.indexOf(el);
        if (index < 0) return;
        uniqueSplice(siblings, index, 0, clonedSelf, parent);
        clones.push(clonedSelf);
      });
      return this.constructor.call(this.constructor, this._makeDomArray(clones));
    };
    exports2.remove = function(selector) {
      var elems = this;
      if (selector)
        elems = elems.filter(selector);
      domEach(elems, function(i, el) {
        var parent = el.parent || el.root;
        if (!parent) {
          return;
        }
        var siblings = parent.children, index = siblings.indexOf(el);
        if (index < 0) return;
        siblings.splice(index, 1);
        if (el.prev) {
          el.prev.next = el.next;
        }
        if (el.next) {
          el.next.prev = el.prev;
        }
        el.prev = el.next = el.parent = el.root = null;
      });
      return this;
    };
    exports2.replaceWith = function(content) {
      var self2 = this;
      domEach(this, function(i, el) {
        var parent = el.parent || el.root;
        if (!parent) {
          return;
        }
        var siblings = parent.children, dom = self2._makeDomArray(typeof content === "function" ? content.call(el, i, el) : content), index;
        updateDOM(dom, null);
        index = siblings.indexOf(el);
        uniqueSplice(siblings, index, 1, dom, parent);
        el.parent = el.prev = el.next = el.root = null;
      });
      return this;
    };
    exports2.empty = function() {
      domEach(this, function(i, el) {
        _.forEach(el.children, function(el2) {
          el2.next = el2.prev = el2.parent = null;
        });
        el.children.length = 0;
      });
      return this;
    };
    exports2.html = function(str) {
      if (str === void 0) {
        if (!this[0] || !this[0].children) return null;
        return $.html(this[0].children, this.options);
      }
      var opts = this.options;
      domEach(this, function(i, el) {
        _.forEach(el.children, function(el2) {
          el2.next = el2.prev = el2.parent = null;
        });
        var content = str.cheerio ? str.clone().get() : evaluate("" + str, opts);
        updateDOM(content, el);
      });
      return this;
    };
    exports2.toString = function() {
      return $.html(this, this.options);
    };
    exports2.text = function(str) {
      if (str === void 0) {
        return $.text(this);
      } else if (typeof str === "function") {
        return domEach(this, function(i, el) {
          var $el = [el];
          return exports2.text.call($el, str.call(el, i, $.text($el)));
        });
      }
      domEach(this, function(i, el) {
        _.forEach(el.children, function(el2) {
          el2.next = el2.prev = el2.parent = null;
        });
        var elem = {
          data: "" + str,
          type: "text",
          parent: el,
          prev: null,
          next: null,
          children: []
        };
        updateDOM(elem, el);
      });
      return this;
    };
    exports2.clone = function() {
      return this._make(cloneDom(this.get(), this.options));
    };
  }
});

// node_modules/lodash.pick/index.js
var require_lodash11 = __commonJS({
  "node_modules/lodash.pick/index.js"(exports2, module2) {
    init_process_shim();
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var symbolTag = "[object Symbol]";
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    function apply(func, thisArg, args) {
      switch (args.length) {
        case 0:
          return func.call(thisArg);
        case 1:
          return func.call(thisArg, args[0]);
        case 2:
          return func.call(thisArg, args[0], args[1]);
        case 3:
          return func.call(thisArg, args[0], args[1], args[2]);
      }
      return func.apply(thisArg, args);
    }
    function arrayMap(array, iteratee) {
      var index = -1, length = array ? array.length : 0, result = Array(length);
      while (++index < length) {
        result[index] = iteratee(array[index], index, array);
      }
      return result;
    }
    function arrayPush(array, values) {
      var index = -1, length = values.length, offset = array.length;
      while (++index < length) {
        array[offset + index] = values[index];
      }
      return array;
    }
    var objectProto = Object.prototype;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var Symbol2 = root.Symbol;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var spreadableSymbol = Symbol2 ? Symbol2.isConcatSpreadable : void 0;
    var nativeMax = Math.max;
    function baseFlatten(array, depth, predicate, isStrict, result) {
      var index = -1, length = array.length;
      predicate || (predicate = isFlattenable);
      result || (result = []);
      while (++index < length) {
        var value = array[index];
        if (depth > 0 && predicate(value)) {
          if (depth > 1) {
            baseFlatten(value, depth - 1, predicate, isStrict, result);
          } else {
            arrayPush(result, value);
          }
        } else if (!isStrict) {
          result[result.length] = value;
        }
      }
      return result;
    }
    function basePick(object, props) {
      object = Object(object);
      return basePickBy(object, props, function(value, key) {
        return key in object;
      });
    }
    function basePickBy(object, props, predicate) {
      var index = -1, length = props.length, result = {};
      while (++index < length) {
        var key = props[index], value = object[key];
        if (predicate(value, key)) {
          result[key] = value;
        }
      }
      return result;
    }
    function baseRest(func, start) {
      start = nativeMax(start === void 0 ? func.length - 1 : start, 0);
      return function() {
        var args = arguments, index = -1, length = nativeMax(args.length - start, 0), array = Array(length);
        while (++index < length) {
          array[index] = args[start + index];
        }
        index = -1;
        var otherArgs = Array(start + 1);
        while (++index < start) {
          otherArgs[index] = args[index];
        }
        otherArgs[start] = array;
        return apply(func, this, otherArgs);
      };
    }
    function isFlattenable(value) {
      return isArray(value) || isArguments(value) || !!(spreadableSymbol && value && value[spreadableSymbol]);
    }
    function toKey(value) {
      if (typeof value == "string" || isSymbol(value)) {
        return value;
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    var pick = baseRest(function(object, props) {
      return object == null ? {} : basePick(object, arrayMap(baseFlatten(props, 1), toKey));
    });
    module2.exports = pick;
  }
});

// node_modules/cheerio/lib/api/css.js
var require_css = __commonJS({
  "node_modules/cheerio/lib/api/css.js"(exports2) {
    init_process_shim();
    var domEach = require_utils().domEach;
    var _ = {
      pick: require_lodash11()
    };
    var toString = Object.prototype.toString;
    exports2.css = function(prop, val) {
      if (arguments.length === 2 || // When `prop` is a "plain" object
      toString.call(prop) === "[object Object]") {
        return domEach(this, function(idx, el) {
          setCss(el, prop, val, idx);
        });
      } else {
        return getCss(this[0], prop);
      }
    };
    function setCss(el, prop, val, idx) {
      if ("string" == typeof prop) {
        var styles = getCss(el);
        if (typeof val === "function") {
          val = val.call(el, idx, styles[prop]);
        }
        if (val === "") {
          delete styles[prop];
        } else if (val != null) {
          styles[prop] = val;
        }
        el.attribs.style = stringify(styles);
      } else if ("object" == typeof prop) {
        Object.keys(prop).forEach(function(k) {
          setCss(el, k, prop[k]);
        });
      }
    }
    function getCss(el, prop) {
      var styles = parse(el.attribs.style);
      if (typeof prop === "string") {
        return styles[prop];
      } else if (Array.isArray(prop)) {
        return _.pick(styles, prop);
      } else {
        return styles;
      }
    }
    function stringify(obj) {
      return Object.keys(obj || {}).reduce(function(str, prop) {
        return str += (str ? " " : "") + prop + ": " + obj[prop] + ";";
      }, "");
    }
    function parse(styles) {
      styles = (styles || "").trim();
      if (!styles) return {};
      return styles.split(";").reduce(function(obj, str) {
        var n = str.indexOf(":");
        if (n < 1 || n === str.length - 1) return obj;
        obj[str.slice(0, n).trim()] = str.slice(n + 1).trim();
        return obj;
      }, {});
    }
  }
});

// node_modules/lodash.map/index.js
var require_lodash12 = __commonJS({
  "node_modules/lodash.map/index.js"(exports2, module2) {
    init_process_shim();
    var LARGE_ARRAY_SIZE = 200;
    var FUNC_ERROR_TEXT = "Expected a function";
    var HASH_UNDEFINED = "__lodash_hash_undefined__";
    var UNORDERED_COMPARE_FLAG = 1;
    var PARTIAL_COMPARE_FLAG = 2;
    var INFINITY = 1 / 0;
    var MAX_SAFE_INTEGER = 9007199254740991;
    var argsTag = "[object Arguments]";
    var arrayTag = "[object Array]";
    var boolTag = "[object Boolean]";
    var dateTag = "[object Date]";
    var errorTag = "[object Error]";
    var funcTag = "[object Function]";
    var genTag = "[object GeneratorFunction]";
    var mapTag = "[object Map]";
    var numberTag = "[object Number]";
    var objectTag = "[object Object]";
    var promiseTag = "[object Promise]";
    var regexpTag = "[object RegExp]";
    var setTag = "[object Set]";
    var stringTag = "[object String]";
    var symbolTag = "[object Symbol]";
    var weakMapTag = "[object WeakMap]";
    var arrayBufferTag = "[object ArrayBuffer]";
    var dataViewTag = "[object DataView]";
    var float32Tag = "[object Float32Array]";
    var float64Tag = "[object Float64Array]";
    var int8Tag = "[object Int8Array]";
    var int16Tag = "[object Int16Array]";
    var int32Tag = "[object Int32Array]";
    var uint8Tag = "[object Uint8Array]";
    var uint8ClampedTag = "[object Uint8ClampedArray]";
    var uint16Tag = "[object Uint16Array]";
    var uint32Tag = "[object Uint32Array]";
    var reIsDeepProp = /\.|\[(?:[^[\]]*|(["'])(?:(?!\1)[^\\]|\\.)*?\1)\]/;
    var reIsPlainProp = /^\w*$/;
    var reLeadingDot = /^\./;
    var rePropName = /[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g;
    var reRegExpChar = /[\\^$.*+?()[\]{}|]/g;
    var reEscapeChar = /\\(\\)?/g;
    var reIsHostCtor = /^\[object .+?Constructor\]$/;
    var reIsUint = /^(?:0|[1-9]\d*)$/;
    var typedArrayTags = {};
    typedArrayTags[float32Tag] = typedArrayTags[float64Tag] = typedArrayTags[int8Tag] = typedArrayTags[int16Tag] = typedArrayTags[int32Tag] = typedArrayTags[uint8Tag] = typedArrayTags[uint8ClampedTag] = typedArrayTags[uint16Tag] = typedArrayTags[uint32Tag] = true;
    typedArrayTags[argsTag] = typedArrayTags[arrayTag] = typedArrayTags[arrayBufferTag] = typedArrayTags[boolTag] = typedArrayTags[dataViewTag] = typedArrayTags[dateTag] = typedArrayTags[errorTag] = typedArrayTags[funcTag] = typedArrayTags[mapTag] = typedArrayTags[numberTag] = typedArrayTags[objectTag] = typedArrayTags[regexpTag] = typedArrayTags[setTag] = typedArrayTags[stringTag] = typedArrayTags[weakMapTag] = false;
    var freeGlobal = typeof globalThis == "object" && globalThis && globalThis.Object === Object && globalThis;
    var freeSelf = typeof self == "object" && self && self.Object === Object && self;
    var root = freeGlobal || freeSelf || Function("return this")();
    var freeExports = typeof exports2 == "object" && exports2 && !exports2.nodeType && exports2;
    var freeModule = freeExports && typeof module2 == "object" && module2 && !module2.nodeType && module2;
    var moduleExports = freeModule && freeModule.exports === freeExports;
    var freeProcess = moduleExports && freeGlobal.process;
    var nodeUtil = (function() {
      try {
        return freeProcess && freeProcess.binding("util");
      } catch (e) {
      }
    })();
    var nodeIsTypedArray = nodeUtil && nodeUtil.isTypedArray;
    function arrayMap(array, iteratee) {
      var index = -1, length = array ? array.length : 0, result = Array(length);
      while (++index < length) {
        result[index] = iteratee(array[index], index, array);
      }
      return result;
    }
    function arraySome(array, predicate) {
      var index = -1, length = array ? array.length : 0;
      while (++index < length) {
        if (predicate(array[index], index, array)) {
          return true;
        }
      }
      return false;
    }
    function baseProperty(key) {
      return function(object) {
        return object == null ? void 0 : object[key];
      };
    }
    function baseTimes(n, iteratee) {
      var index = -1, result = Array(n);
      while (++index < n) {
        result[index] = iteratee(index);
      }
      return result;
    }
    function baseUnary(func) {
      return function(value) {
        return func(value);
      };
    }
    function getValue(object, key) {
      return object == null ? void 0 : object[key];
    }
    function isHostObject(value) {
      var result = false;
      if (value != null && typeof value.toString != "function") {
        try {
          result = !!(value + "");
        } catch (e) {
        }
      }
      return result;
    }
    function mapToArray(map2) {
      var index = -1, result = Array(map2.size);
      map2.forEach(function(value, key) {
        result[++index] = [key, value];
      });
      return result;
    }
    function overArg(func, transform) {
      return function(arg) {
        return func(transform(arg));
      };
    }
    function setToArray(set) {
      var index = -1, result = Array(set.size);
      set.forEach(function(value) {
        result[++index] = value;
      });
      return result;
    }
    var arrayProto = Array.prototype;
    var funcProto = Function.prototype;
    var objectProto = Object.prototype;
    var coreJsData = root["__core-js_shared__"];
    var maskSrcKey = (function() {
      var uid = /[^.]+$/.exec(coreJsData && coreJsData.keys && coreJsData.keys.IE_PROTO || "");
      return uid ? "Symbol(src)_1." + uid : "";
    })();
    var funcToString = funcProto.toString;
    var hasOwnProperty2 = objectProto.hasOwnProperty;
    var objectToString = objectProto.toString;
    var reIsNative = RegExp(
      "^" + funcToString.call(hasOwnProperty2).replace(reRegExpChar, "\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g, "$1.*?") + "$"
    );
    var Symbol2 = root.Symbol;
    var Uint8Array2 = root.Uint8Array;
    var propertyIsEnumerable = objectProto.propertyIsEnumerable;
    var splice = arrayProto.splice;
    var nativeKeys = overArg(Object.keys, Object);
    var DataView = getNative(root, "DataView");
    var Map = getNative(root, "Map");
    var Promise2 = getNative(root, "Promise");
    var Set = getNative(root, "Set");
    var WeakMap = getNative(root, "WeakMap");
    var nativeCreate = getNative(Object, "create");
    var dataViewCtorString = toSource(DataView);
    var mapCtorString = toSource(Map);
    var promiseCtorString = toSource(Promise2);
    var setCtorString = toSource(Set);
    var weakMapCtorString = toSource(WeakMap);
    var symbolProto = Symbol2 ? Symbol2.prototype : void 0;
    var symbolValueOf = symbolProto ? symbolProto.valueOf : void 0;
    var symbolToString = symbolProto ? symbolProto.toString : void 0;
    function Hash(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function hashClear() {
      this.__data__ = nativeCreate ? nativeCreate(null) : {};
    }
    function hashDelete(key) {
      return this.has(key) && delete this.__data__[key];
    }
    function hashGet(key) {
      var data = this.__data__;
      if (nativeCreate) {
        var result = data[key];
        return result === HASH_UNDEFINED ? void 0 : result;
      }
      return hasOwnProperty2.call(data, key) ? data[key] : void 0;
    }
    function hashHas(key) {
      var data = this.__data__;
      return nativeCreate ? data[key] !== void 0 : hasOwnProperty2.call(data, key);
    }
    function hashSet(key, value) {
      var data = this.__data__;
      data[key] = nativeCreate && value === void 0 ? HASH_UNDEFINED : value;
      return this;
    }
    Hash.prototype.clear = hashClear;
    Hash.prototype["delete"] = hashDelete;
    Hash.prototype.get = hashGet;
    Hash.prototype.has = hashHas;
    Hash.prototype.set = hashSet;
    function ListCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function listCacheClear() {
      this.__data__ = [];
    }
    function listCacheDelete(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        return false;
      }
      var lastIndex = data.length - 1;
      if (index == lastIndex) {
        data.pop();
      } else {
        splice.call(data, index, 1);
      }
      return true;
    }
    function listCacheGet(key) {
      var data = this.__data__, index = assocIndexOf(data, key);
      return index < 0 ? void 0 : data[index][1];
    }
    function listCacheHas(key) {
      return assocIndexOf(this.__data__, key) > -1;
    }
    function listCacheSet(key, value) {
      var data = this.__data__, index = assocIndexOf(data, key);
      if (index < 0) {
        data.push([key, value]);
      } else {
        data[index][1] = value;
      }
      return this;
    }
    ListCache.prototype.clear = listCacheClear;
    ListCache.prototype["delete"] = listCacheDelete;
    ListCache.prototype.get = listCacheGet;
    ListCache.prototype.has = listCacheHas;
    ListCache.prototype.set = listCacheSet;
    function MapCache(entries) {
      var index = -1, length = entries ? entries.length : 0;
      this.clear();
      while (++index < length) {
        var entry = entries[index];
        this.set(entry[0], entry[1]);
      }
    }
    function mapCacheClear() {
      this.__data__ = {
        "hash": new Hash(),
        "map": new (Map || ListCache)(),
        "string": new Hash()
      };
    }
    function mapCacheDelete(key) {
      return getMapData(this, key)["delete"](key);
    }
    function mapCacheGet(key) {
      return getMapData(this, key).get(key);
    }
    function mapCacheHas(key) {
      return getMapData(this, key).has(key);
    }
    function mapCacheSet(key, value) {
      getMapData(this, key).set(key, value);
      return this;
    }
    MapCache.prototype.clear = mapCacheClear;
    MapCache.prototype["delete"] = mapCacheDelete;
    MapCache.prototype.get = mapCacheGet;
    MapCache.prototype.has = mapCacheHas;
    MapCache.prototype.set = mapCacheSet;
    function SetCache(values) {
      var index = -1, length = values ? values.length : 0;
      this.__data__ = new MapCache();
      while (++index < length) {
        this.add(values[index]);
      }
    }
    function setCacheAdd(value) {
      this.__data__.set(value, HASH_UNDEFINED);
      return this;
    }
    function setCacheHas(value) {
      return this.__data__.has(value);
    }
    SetCache.prototype.add = SetCache.prototype.push = setCacheAdd;
    SetCache.prototype.has = setCacheHas;
    function Stack(entries) {
      this.__data__ = new ListCache(entries);
    }
    function stackClear() {
      this.__data__ = new ListCache();
    }
    function stackDelete(key) {
      return this.__data__["delete"](key);
    }
    function stackGet(key) {
      return this.__data__.get(key);
    }
    function stackHas(key) {
      return this.__data__.has(key);
    }
    function stackSet(key, value) {
      var cache = this.__data__;
      if (cache instanceof ListCache) {
        var pairs = cache.__data__;
        if (!Map || pairs.length < LARGE_ARRAY_SIZE - 1) {
          pairs.push([key, value]);
          return this;
        }
        cache = this.__data__ = new MapCache(pairs);
      }
      cache.set(key, value);
      return this;
    }
    Stack.prototype.clear = stackClear;
    Stack.prototype["delete"] = stackDelete;
    Stack.prototype.get = stackGet;
    Stack.prototype.has = stackHas;
    Stack.prototype.set = stackSet;
    function arrayLikeKeys(value, inherited) {
      var result = isArray(value) || isArguments(value) ? baseTimes(value.length, String) : [];
      var length = result.length, skipIndexes = !!length;
      for (var key in value) {
        if ((inherited || hasOwnProperty2.call(value, key)) && !(skipIndexes && (key == "length" || isIndex(key, length)))) {
          result.push(key);
        }
      }
      return result;
    }
    function assocIndexOf(array, key) {
      var length = array.length;
      while (length--) {
        if (eq(array[length][0], key)) {
          return length;
        }
      }
      return -1;
    }
    var baseEach = createBaseEach(baseForOwn);
    var baseFor = createBaseFor();
    function baseForOwn(object, iteratee) {
      return object && baseFor(object, iteratee, keys);
    }
    function baseGet(object, path) {
      path = isKey(path, object) ? [path] : castPath(path);
      var index = 0, length = path.length;
      while (object != null && index < length) {
        object = object[toKey(path[index++])];
      }
      return index && index == length ? object : void 0;
    }
    function baseGetTag(value) {
      return objectToString.call(value);
    }
    function baseHasIn(object, key) {
      return object != null && key in Object(object);
    }
    function baseIsEqual(value, other, customizer, bitmask, stack) {
      if (value === other) {
        return true;
      }
      if (value == null || other == null || !isObject(value) && !isObjectLike(other)) {
        return value !== value && other !== other;
      }
      return baseIsEqualDeep(value, other, baseIsEqual, customizer, bitmask, stack);
    }
    function baseIsEqualDeep(object, other, equalFunc, customizer, bitmask, stack) {
      var objIsArr = isArray(object), othIsArr = isArray(other), objTag = arrayTag, othTag = arrayTag;
      if (!objIsArr) {
        objTag = getTag(object);
        objTag = objTag == argsTag ? objectTag : objTag;
      }
      if (!othIsArr) {
        othTag = getTag(other);
        othTag = othTag == argsTag ? objectTag : othTag;
      }
      var objIsObj = objTag == objectTag && !isHostObject(object), othIsObj = othTag == objectTag && !isHostObject(other), isSameTag = objTag == othTag;
      if (isSameTag && !objIsObj) {
        stack || (stack = new Stack());
        return objIsArr || isTypedArray(object) ? equalArrays(object, other, equalFunc, customizer, bitmask, stack) : equalByTag(object, other, objTag, equalFunc, customizer, bitmask, stack);
      }
      if (!(bitmask & PARTIAL_COMPARE_FLAG)) {
        var objIsWrapped = objIsObj && hasOwnProperty2.call(object, "__wrapped__"), othIsWrapped = othIsObj && hasOwnProperty2.call(other, "__wrapped__");
        if (objIsWrapped || othIsWrapped) {
          var objUnwrapped = objIsWrapped ? object.value() : object, othUnwrapped = othIsWrapped ? other.value() : other;
          stack || (stack = new Stack());
          return equalFunc(objUnwrapped, othUnwrapped, customizer, bitmask, stack);
        }
      }
      if (!isSameTag) {
        return false;
      }
      stack || (stack = new Stack());
      return equalObjects(object, other, equalFunc, customizer, bitmask, stack);
    }
    function baseIsMatch(object, source, matchData, customizer) {
      var index = matchData.length, length = index, noCustomizer = !customizer;
      if (object == null) {
        return !length;
      }
      object = Object(object);
      while (index--) {
        var data = matchData[index];
        if (noCustomizer && data[2] ? data[1] !== object[data[0]] : !(data[0] in object)) {
          return false;
        }
      }
      while (++index < length) {
        data = matchData[index];
        var key = data[0], objValue = object[key], srcValue = data[1];
        if (noCustomizer && data[2]) {
          if (objValue === void 0 && !(key in object)) {
            return false;
          }
        } else {
          var stack = new Stack();
          if (customizer) {
            var result = customizer(objValue, srcValue, key, object, source, stack);
          }
          if (!(result === void 0 ? baseIsEqual(srcValue, objValue, customizer, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG, stack) : result)) {
            return false;
          }
        }
      }
      return true;
    }
    function baseIsNative(value) {
      if (!isObject(value) || isMasked(value)) {
        return false;
      }
      var pattern = isFunction(value) || isHostObject(value) ? reIsNative : reIsHostCtor;
      return pattern.test(toSource(value));
    }
    function baseIsTypedArray(value) {
      return isObjectLike(value) && isLength(value.length) && !!typedArrayTags[objectToString.call(value)];
    }
    function baseIteratee(value) {
      if (typeof value == "function") {
        return value;
      }
      if (value == null) {
        return identity;
      }
      if (typeof value == "object") {
        return isArray(value) ? baseMatchesProperty(value[0], value[1]) : baseMatches(value);
      }
      return property(value);
    }
    function baseKeys(object) {
      if (!isPrototype(object)) {
        return nativeKeys(object);
      }
      var result = [];
      for (var key in Object(object)) {
        if (hasOwnProperty2.call(object, key) && key != "constructor") {
          result.push(key);
        }
      }
      return result;
    }
    function baseMap(collection, iteratee) {
      var index = -1, result = isArrayLike(collection) ? Array(collection.length) : [];
      baseEach(collection, function(value, key, collection2) {
        result[++index] = iteratee(value, key, collection2);
      });
      return result;
    }
    function baseMatches(source) {
      var matchData = getMatchData(source);
      if (matchData.length == 1 && matchData[0][2]) {
        return matchesStrictComparable(matchData[0][0], matchData[0][1]);
      }
      return function(object) {
        return object === source || baseIsMatch(object, source, matchData);
      };
    }
    function baseMatchesProperty(path, srcValue) {
      if (isKey(path) && isStrictComparable(srcValue)) {
        return matchesStrictComparable(toKey(path), srcValue);
      }
      return function(object) {
        var objValue = get(object, path);
        return objValue === void 0 && objValue === srcValue ? hasIn(object, path) : baseIsEqual(srcValue, objValue, void 0, UNORDERED_COMPARE_FLAG | PARTIAL_COMPARE_FLAG);
      };
    }
    function basePropertyDeep(path) {
      return function(object) {
        return baseGet(object, path);
      };
    }
    function baseToString(value) {
      if (typeof value == "string") {
        return value;
      }
      if (isSymbol(value)) {
        return symbolToString ? symbolToString.call(value) : "";
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function castPath(value) {
      return isArray(value) ? value : stringToPath(value);
    }
    function createBaseEach(eachFunc, fromRight) {
      return function(collection, iteratee) {
        if (collection == null) {
          return collection;
        }
        if (!isArrayLike(collection)) {
          return eachFunc(collection, iteratee);
        }
        var length = collection.length, index = fromRight ? length : -1, iterable = Object(collection);
        while (fromRight ? index-- : ++index < length) {
          if (iteratee(iterable[index], index, iterable) === false) {
            break;
          }
        }
        return collection;
      };
    }
    function createBaseFor(fromRight) {
      return function(object, iteratee, keysFunc) {
        var index = -1, iterable = Object(object), props = keysFunc(object), length = props.length;
        while (length--) {
          var key = props[fromRight ? length : ++index];
          if (iteratee(iterable[key], key, iterable) === false) {
            break;
          }
        }
        return object;
      };
    }
    function equalArrays(array, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, arrLength = array.length, othLength = other.length;
      if (arrLength != othLength && !(isPartial && othLength > arrLength)) {
        return false;
      }
      var stacked = stack.get(array);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var index = -1, result = true, seen = bitmask & UNORDERED_COMPARE_FLAG ? new SetCache() : void 0;
      stack.set(array, other);
      stack.set(other, array);
      while (++index < arrLength) {
        var arrValue = array[index], othValue = other[index];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, arrValue, index, other, array, stack) : customizer(arrValue, othValue, index, array, other, stack);
        }
        if (compared !== void 0) {
          if (compared) {
            continue;
          }
          result = false;
          break;
        }
        if (seen) {
          if (!arraySome(other, function(othValue2, othIndex) {
            if (!seen.has(othIndex) && (arrValue === othValue2 || equalFunc(arrValue, othValue2, customizer, bitmask, stack))) {
              return seen.add(othIndex);
            }
          })) {
            result = false;
            break;
          }
        } else if (!(arrValue === othValue || equalFunc(arrValue, othValue, customizer, bitmask, stack))) {
          result = false;
          break;
        }
      }
      stack["delete"](array);
      stack["delete"](other);
      return result;
    }
    function equalByTag(object, other, tag, equalFunc, customizer, bitmask, stack) {
      switch (tag) {
        case dataViewTag:
          if (object.byteLength != other.byteLength || object.byteOffset != other.byteOffset) {
            return false;
          }
          object = object.buffer;
          other = other.buffer;
        case arrayBufferTag:
          if (object.byteLength != other.byteLength || !equalFunc(new Uint8Array2(object), new Uint8Array2(other))) {
            return false;
          }
          return true;
        case boolTag:
        case dateTag:
        case numberTag:
          return eq(+object, +other);
        case errorTag:
          return object.name == other.name && object.message == other.message;
        case regexpTag:
        case stringTag:
          return object == other + "";
        case mapTag:
          var convert = mapToArray;
        case setTag:
          var isPartial = bitmask & PARTIAL_COMPARE_FLAG;
          convert || (convert = setToArray);
          if (object.size != other.size && !isPartial) {
            return false;
          }
          var stacked = stack.get(object);
          if (stacked) {
            return stacked == other;
          }
          bitmask |= UNORDERED_COMPARE_FLAG;
          stack.set(object, other);
          var result = equalArrays(convert(object), convert(other), equalFunc, customizer, bitmask, stack);
          stack["delete"](object);
          return result;
        case symbolTag:
          if (symbolValueOf) {
            return symbolValueOf.call(object) == symbolValueOf.call(other);
          }
      }
      return false;
    }
    function equalObjects(object, other, equalFunc, customizer, bitmask, stack) {
      var isPartial = bitmask & PARTIAL_COMPARE_FLAG, objProps = keys(object), objLength = objProps.length, othProps = keys(other), othLength = othProps.length;
      if (objLength != othLength && !isPartial) {
        return false;
      }
      var index = objLength;
      while (index--) {
        var key = objProps[index];
        if (!(isPartial ? key in other : hasOwnProperty2.call(other, key))) {
          return false;
        }
      }
      var stacked = stack.get(object);
      if (stacked && stack.get(other)) {
        return stacked == other;
      }
      var result = true;
      stack.set(object, other);
      stack.set(other, object);
      var skipCtor = isPartial;
      while (++index < objLength) {
        key = objProps[index];
        var objValue = object[key], othValue = other[key];
        if (customizer) {
          var compared = isPartial ? customizer(othValue, objValue, key, other, object, stack) : customizer(objValue, othValue, key, object, other, stack);
        }
        if (!(compared === void 0 ? objValue === othValue || equalFunc(objValue, othValue, customizer, bitmask, stack) : compared)) {
          result = false;
          break;
        }
        skipCtor || (skipCtor = key == "constructor");
      }
      if (result && !skipCtor) {
        var objCtor = object.constructor, othCtor = other.constructor;
        if (objCtor != othCtor && ("constructor" in object && "constructor" in other) && !(typeof objCtor == "function" && objCtor instanceof objCtor && typeof othCtor == "function" && othCtor instanceof othCtor)) {
          result = false;
        }
      }
      stack["delete"](object);
      stack["delete"](other);
      return result;
    }
    function getMapData(map2, key) {
      var data = map2.__data__;
      return isKeyable(key) ? data[typeof key == "string" ? "string" : "hash"] : data.map;
    }
    function getMatchData(object) {
      var result = keys(object), length = result.length;
      while (length--) {
        var key = result[length], value = object[key];
        result[length] = [key, value, isStrictComparable(value)];
      }
      return result;
    }
    function getNative(object, key) {
      var value = getValue(object, key);
      return baseIsNative(value) ? value : void 0;
    }
    var getTag = baseGetTag;
    if (DataView && getTag(new DataView(new ArrayBuffer(1))) != dataViewTag || Map && getTag(new Map()) != mapTag || Promise2 && getTag(Promise2.resolve()) != promiseTag || Set && getTag(new Set()) != setTag || WeakMap && getTag(new WeakMap()) != weakMapTag) {
      getTag = function(value) {
        var result = objectToString.call(value), Ctor = result == objectTag ? value.constructor : void 0, ctorString = Ctor ? toSource(Ctor) : void 0;
        if (ctorString) {
          switch (ctorString) {
            case dataViewCtorString:
              return dataViewTag;
            case mapCtorString:
              return mapTag;
            case promiseCtorString:
              return promiseTag;
            case setCtorString:
              return setTag;
            case weakMapCtorString:
              return weakMapTag;
          }
        }
        return result;
      };
    }
    function hasPath(object, path, hasFunc) {
      path = isKey(path, object) ? [path] : castPath(path);
      var result, index = -1, length = path.length;
      while (++index < length) {
        var key = toKey(path[index]);
        if (!(result = object != null && hasFunc(object, key))) {
          break;
        }
        object = object[key];
      }
      if (result) {
        return result;
      }
      var length = object ? object.length : 0;
      return !!length && isLength(length) && isIndex(key, length) && (isArray(object) || isArguments(object));
    }
    function isIndex(value, length) {
      length = length == null ? MAX_SAFE_INTEGER : length;
      return !!length && (typeof value == "number" || reIsUint.test(value)) && (value > -1 && value % 1 == 0 && value < length);
    }
    function isKey(value, object) {
      if (isArray(value)) {
        return false;
      }
      var type = typeof value;
      if (type == "number" || type == "symbol" || type == "boolean" || value == null || isSymbol(value)) {
        return true;
      }
      return reIsPlainProp.test(value) || !reIsDeepProp.test(value) || object != null && value in Object(object);
    }
    function isKeyable(value) {
      var type = typeof value;
      return type == "string" || type == "number" || type == "symbol" || type == "boolean" ? value !== "__proto__" : value === null;
    }
    function isMasked(func) {
      return !!maskSrcKey && maskSrcKey in func;
    }
    function isPrototype(value) {
      var Ctor = value && value.constructor, proto = typeof Ctor == "function" && Ctor.prototype || objectProto;
      return value === proto;
    }
    function isStrictComparable(value) {
      return value === value && !isObject(value);
    }
    function matchesStrictComparable(key, srcValue) {
      return function(object) {
        if (object == null) {
          return false;
        }
        return object[key] === srcValue && (srcValue !== void 0 || key in Object(object));
      };
    }
    var stringToPath = memoize(function(string) {
      string = toString(string);
      var result = [];
      if (reLeadingDot.test(string)) {
        result.push("");
      }
      string.replace(rePropName, function(match, number, quote, string2) {
        result.push(quote ? string2.replace(reEscapeChar, "$1") : number || match);
      });
      return result;
    });
    function toKey(value) {
      if (typeof value == "string" || isSymbol(value)) {
        return value;
      }
      var result = value + "";
      return result == "0" && 1 / value == -INFINITY ? "-0" : result;
    }
    function toSource(func) {
      if (func != null) {
        try {
          return funcToString.call(func);
        } catch (e) {
        }
        try {
          return func + "";
        } catch (e) {
        }
      }
      return "";
    }
    function map(collection, iteratee) {
      var func = isArray(collection) ? arrayMap : baseMap;
      return func(collection, baseIteratee(iteratee, 3));
    }
    function memoize(func, resolver) {
      if (typeof func != "function" || resolver && typeof resolver != "function") {
        throw new TypeError(FUNC_ERROR_TEXT);
      }
      var memoized = function() {
        var args = arguments, key = resolver ? resolver.apply(this, args) : args[0], cache = memoized.cache;
        if (cache.has(key)) {
          return cache.get(key);
        }
        var result = func.apply(this, args);
        memoized.cache = cache.set(key, result);
        return result;
      };
      memoized.cache = new (memoize.Cache || MapCache)();
      return memoized;
    }
    memoize.Cache = MapCache;
    function eq(value, other) {
      return value === other || value !== value && other !== other;
    }
    function isArguments(value) {
      return isArrayLikeObject(value) && hasOwnProperty2.call(value, "callee") && (!propertyIsEnumerable.call(value, "callee") || objectToString.call(value) == argsTag);
    }
    var isArray = Array.isArray;
    function isArrayLike(value) {
      return value != null && isLength(value.length) && !isFunction(value);
    }
    function isArrayLikeObject(value) {
      return isObjectLike(value) && isArrayLike(value);
    }
    function isFunction(value) {
      var tag = isObject(value) ? objectToString.call(value) : "";
      return tag == funcTag || tag == genTag;
    }
    function isLength(value) {
      return typeof value == "number" && value > -1 && value % 1 == 0 && value <= MAX_SAFE_INTEGER;
    }
    function isObject(value) {
      var type = typeof value;
      return !!value && (type == "object" || type == "function");
    }
    function isObjectLike(value) {
      return !!value && typeof value == "object";
    }
    function isSymbol(value) {
      return typeof value == "symbol" || isObjectLike(value) && objectToString.call(value) == symbolTag;
    }
    var isTypedArray = nodeIsTypedArray ? baseUnary(nodeIsTypedArray) : baseIsTypedArray;
    function toString(value) {
      return value == null ? "" : baseToString(value);
    }
    function get(object, path, defaultValue) {
      var result = object == null ? void 0 : baseGet(object, path);
      return result === void 0 ? defaultValue : result;
    }
    function hasIn(object, path) {
      return object != null && hasPath(object, path, baseHasIn);
    }
    function keys(object) {
      return isArrayLike(object) ? arrayLikeKeys(object) : baseKeys(object);
    }
    function identity(value) {
      return value;
    }
    function property(path) {
      return isKey(path) ? baseProperty(toKey(path)) : basePropertyDeep(path);
    }
    module2.exports = map;
  }
});

// node_modules/cheerio/lib/api/forms.js
var require_forms = __commonJS({
  "node_modules/cheerio/lib/api/forms.js"(exports2) {
    init_process_shim();
    var submittableSelector = "input,select,textarea,keygen";
    var r20 = /%20/g;
    var rCRLF = /\r?\n/g;
    var _ = {
      map: require_lodash12()
    };
    exports2.serialize = function() {
      var arr = this.serializeArray();
      var retArr = _.map(arr, function(data) {
        return encodeURIComponent(data.name) + "=" + encodeURIComponent(data.value);
      });
      return retArr.join("&").replace(r20, "+");
    };
    exports2.serializeArray = function() {
      var Cheerio = this.constructor;
      return this.map(function() {
        var elem = this;
        var $elem = Cheerio(elem);
        if (elem.name === "form") {
          return $elem.find(submittableSelector).toArray();
        } else {
          return $elem.filter(submittableSelector).toArray();
        }
      }).filter(
        // Verify elements have a name (`attr.name`) and are not disabled (`:disabled`)
        '[name!=""]:not(:disabled):not(:submit, :button, :image, :reset, :file):matches([checked], :not(:checkbox, :radio))'
        // Convert each of the elements to its value(s)
      ).map(function(i, elem) {
        var $elem = Cheerio(elem);
        var name = $elem.attr("name");
        var val = $elem.val();
        if (val == null) {
          return null;
        } else {
          if (Array.isArray(val)) {
            return _.map(val, function(val2) {
              return { name, value: val2.replace(rCRLF, "\r\n") };
            });
          } else {
            return { name, value: val.replace(rCRLF, "\r\n") };
          }
        }
      }).get();
    };
  }
});

// node_modules/cheerio/lib/cheerio.js
var require_cheerio = __commonJS({
  "node_modules/cheerio/lib/cheerio.js"(exports2, module2) {
    init_process_shim();
    var parse = require_parse();
    var isHtml = require_utils().isHtml;
    var _ = {
      extend: require_lodash(),
      bind: require_lodash2(),
      forEach: require_lodash3(),
      defaults: require_lodash4()
    };
    var api = [
      require_attributes2(),
      require_traversing(),
      require_manipulation2(),
      require_css(),
      require_forms()
    ];
    var Cheerio = module2.exports = function(selector, context, root, options) {
      if (!(this instanceof Cheerio)) return new Cheerio(selector, context, root, options);
      this.options = _.defaults(options || {}, this.options);
      if (!selector) return this;
      if (root) {
        if (typeof root === "string") root = parse(root, this.options);
        this._root = Cheerio.call(this, root);
      }
      if (selector.cheerio) return selector;
      if (isNode(selector))
        selector = [selector];
      if (Array.isArray(selector)) {
        _.forEach(selector, _.bind(function(elem, idx) {
          this[idx] = elem;
        }, this));
        this.length = selector.length;
        return this;
      }
      if (typeof selector === "string" && isHtml(selector)) {
        return Cheerio.call(this, parse(selector, this.options).children);
      }
      if (!context) {
        context = this._root;
      } else if (typeof context === "string") {
        if (isHtml(context)) {
          context = parse(context, this.options);
          context = Cheerio.call(this, context);
        } else {
          selector = [context, selector].join(" ");
          context = this._root;
        }
      } else if (!context.cheerio) {
        context = Cheerio.call(this, context);
      }
      if (!context) return this;
      return context.find(selector);
    };
    _.extend(Cheerio, require_static());
    Cheerio.prototype.cheerio = "[cheerio object]";
    Cheerio.prototype.options = {
      withDomLvl1: true,
      normalizeWhitespace: false,
      xmlMode: false,
      decodeEntities: true
    };
    Cheerio.prototype.length = 0;
    Cheerio.prototype.splice = Array.prototype.splice;
    Cheerio.prototype._make = function(dom, context) {
      var cheerio = new this.constructor(dom, context, this._root, this.options);
      cheerio.prevObject = this;
      return cheerio;
    };
    Cheerio.prototype.toArray = function() {
      return this.get();
    };
    api.forEach(function(mod) {
      _.extend(Cheerio.prototype, mod);
    });
    var isNode = function(obj) {
      return obj.name || obj.type === "text" || obj.type === "comment";
    };
  }
});

// node_modules/cheerio/package.json
var require_package = __commonJS({
  "node_modules/cheerio/package.json"(exports2, module2) {
    module2.exports = {
      name: "cheerio",
      version: "0.22.0",
      description: "Tiny, fast, and elegant implementation of core jQuery designed specifically for the server",
      author: "Matt Mueller <mattmuelle@gmail.com> (mat.io)",
      license: "MIT",
      keywords: [
        "htmlparser",
        "jquery",
        "selector",
        "scraper",
        "parser",
        "html"
      ],
      repository: {
        type: "git",
        url: "git://github.com/cheeriojs/cheerio.git"
      },
      main: "./index.js",
      files: [
        "index.js",
        "lib"
      ],
      engines: {
        node: ">= 0.6"
      },
      dependencies: {
        "css-select": "~1.2.0",
        "dom-serializer": "~0.1.0",
        entities: "~1.1.1",
        htmlparser2: "^3.9.1",
        "lodash.assignin": "^4.0.9",
        "lodash.bind": "^4.1.4",
        "lodash.defaults": "^4.0.1",
        "lodash.filter": "^4.4.0",
        "lodash.flatten": "^4.2.0",
        "lodash.foreach": "^4.3.0",
        "lodash.map": "^4.4.0",
        "lodash.merge": "^4.4.0",
        "lodash.pick": "^4.2.1",
        "lodash.reduce": "^4.4.0",
        "lodash.reject": "^4.4.0",
        "lodash.some": "^4.4.0"
      },
      devDependencies: {
        benchmark: "^2.1.0",
        coveralls: "^2.11.9",
        "expect.js": "~0.3.1",
        istanbul: "^0.4.3",
        jsdom: "^9.2.1",
        jquery: "^3.0.0",
        jshint: "^2.9.2",
        mocha: "^2.5.3",
        xyz: "~0.5.0"
      },
      scripts: {
        test: "make test"
      }
    };
  }
});

// node_modules/cheerio/index.js
init_process_shim();
exports = module.exports = require_cheerio();
exports.version = require_package().version;
