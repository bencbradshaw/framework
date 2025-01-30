export function html(strings: TemplateStringsArray, ...values: any[]) {
  return strings.reduce((result, string, i) => {
    const value = values[i] !== undefined ? values[i] : '';
    return result + string + value;
  }, '');
}
