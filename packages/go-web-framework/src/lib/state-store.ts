// create a decorator for getter setter pair that can also emit an event
export function prop() {
  return function (target: EventTarget, key: string) {
    let value = (target as any)[key];
    Object.defineProperty(target, key, {
      get() {
        return value;
      },
      set(newValue) {
        value = newValue;
        this.dispatchEvent(new CustomEvent(key, { detail: value, bubbles: true }));
      }
    });
  };
}

export class StateStore extends EventTarget {
  subscribe<K extends keyof this>(key: K, cb: (value: this[K]) => void) {
    const value = this[key];
    if (value !== undefined && value !== null) {
      cb(value);
    }
    const eventListener = (event: Event) => {
      cb(this[key]);
    };
    this.addEventListener(key as string, eventListener);
    return {
      unsubscribe: () => {
        this.removeEventListener(key as string, eventListener);
      }
    };
  }
}
