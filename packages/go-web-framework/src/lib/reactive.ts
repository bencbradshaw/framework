export function reactive() {
  return function (target: any, key: string) {
    const privateKey = Symbol(key);

    Object.defineProperty(target, key, {
      get() {
        return this[privateKey];
      },
      set(value: any) {
        this[privateKey] = value;
        if (this.shadowRoot && this.update && typeof this.update === 'function') {
          if (!this.isUpdateScheduled) {
            this.isUpdateScheduled = true;
            Promise.resolve().then(() => {
              this.update();
              this.isUpdateScheduled = false;
            });
          }
        }
      },
      enumerable: true,
      configurable: true
    });
  };
}
