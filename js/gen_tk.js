#!/usr/bin/env nodejs

// TODO: get TKK from https://translate.google.com/

//TKK=eval('((function(){var a\x3d1944843558;var b\x3d2099926662;return 410250+\x27.\x27+(a+b)})())');

var TKK = function(x,a,b) {
    return x+'.'+(a+b);
}

var fj = function(a) {
        return function() {
            return a
        }
    },
    hj = null;
    gj = function(a, b) {
        for (var c = 0; c < b.length - 2; c += 3) {
            var d = b.charAt(c + 2),
                d = "a" <= d ? d.charCodeAt(0) - 87 : Number(d),
                d = "+" == b.charAt(c + 1) ? a >>> d : a << d;
            a = "+" == b.charAt(c) ? a + d & 4294967295 : a ^ d
        }
        return a
    },

    ij = function(a,tkkx,tkka,tkkb) {
        var b;
        if (null !== hj) b = hj;
        else {
            b = fj(String.fromCharCode(84));
            var c = fj(String.fromCharCode(75));
            b = [b(), b()];
            b[1] = c();
            b = (hj = TKK(tkkx,tkka,tkkb) || "") || ""
        }
        var d = fj(String.fromCharCode(116)),
            c = fj(String.fromCharCode(107)),
            d = [d(), d()];
        d[1] = c();
        c = "&" + d.join("") +
            "=";
        d = b.split(".");
        b = Number(d[0]) || 0;
        for (var e = [], f = 0, k = 0; k < a.length; k++) {
            var l = a.charCodeAt(k);
            128 > l ? e[f++] = l : (2048 > l ? e[f++] = l >> 6 | 192 : (55296 == (l & 64512) && k + 1 < a.length && 56320 == (a.charCodeAt(k + 1) & 64512) ? (l = 65536 + ((l & 1023) << 10) + (a.charCodeAt(++k) & 1023), e[f++] = l >> 18 | 240, e[f++] = l >> 12 & 63 | 128) : e[f++] = l >> 12 | 224, e[f++] = l >> 6 & 63 | 128), e[f++] = l & 63 | 128)
        }
        a = b;
        for (f = 0; f < e.length; f++) a += e[f], a = gj(a, "+-a^+6");
        a = gj(a, "+-3^+b+-f");
        a ^= Number(d[1]) || 0;
        0 > a && (a = (a & 2147483647) + 2147483648);
        a %= 1E6;
        return (a.toString() + "." +
            (a ^ b))
    };

if (process.argv.length === 6) {
	console.log(ij(process.argv[2],parseInt(process.argv[3]),parseInt(process.argv[4]),parseInt(process.argv[5])));
}
