function encodeUri(uri) {
    return encodeURI(uri)
}


var keyStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";

function getEncode(dataStr, userAccount, userPassword) {
    if (dataStr == "no") {
        return false;
    } else {
        var scode = dataStr.split("#")[0];
        var sxh = dataStr.split("#")[1];
        var code = userAccount + "%%%" + userPassword;
        var encoded = "";
        for (var i = 0; i < code.length; i++) {
            if (i < 20) {
                encoded = encoded + code.substring(i, i + 1) + scode.substring(0, parseInt(sxh.substring(i, i + 1)));
                scode = scode.substring(parseInt(sxh.substring(i, i + 1)), scode.length);
            } else {
                encoded = encoded + code.substring(i, code.length);
                i = code.length;
            }
        }
    }
    return encoded
}


function encodeInp(input) {
    var output = "";
    var chr1, chr2, chr3 = "";
    var enc1, enc2, enc3, enc4 = "";
    var i = 0;
    do {
        chr1 = input.charCodeAt(i++);
        chr2 = input.charCodeAt(i++);
        chr3 = input.charCodeAt(i++);
        enc1 = chr1 >> 2;
        enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
        enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
        enc4 = chr3 & 63;
        if (isNaN(chr2)) {
            enc3 = enc4 = 64
        } else if (isNaN(chr3)) {
            enc4 = 64
        }
        output = output + keyStr.charAt(enc1) + keyStr.charAt(enc2) + keyStr.charAt(enc3) + keyStr.charAt(enc4);
        chr1 = chr2 = chr3 = "";
        enc1 = enc2 = enc3 = enc4 = ""
    } while (i < input.length);
    return output
}
