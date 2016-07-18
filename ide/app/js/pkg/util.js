export default {
  splitN(str, sep, n) {
    if (!n) {
      return [];
    } else if (n === 1) {
      return [str];
    } else if (n < 0) {
      return str.split(sep);
    }
    const arr = [];
    while (n > 1) {
      const i = str.indexOf(sep);
      arr.push(str.substring(0, i));
      str = str.substring(i+1);
      n--;
    }
    arr.push(str);
    return arr;
  }
};
