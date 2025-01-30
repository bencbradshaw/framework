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
        console.log('THIS', this);
        this.dispatchEvent(new CustomEvent(key, { detail: value, bubbles: true }));
      }
    });
  };
}

export class StateStore extends EventTarget {
  subscribe(key: any, callback: (value: any) => void) {
    callback((this as any)[key]);
    const listenerHandler = (event: CustomEvent<any>) => {
      console.log('event listener handler', event);
      callback(event.detail);
    };
    this.addEventListener(key, listenerHandler as EventListener);
    return {
      unsubscribe: this.removeEventListener(key, listenerHandler as EventListener)
    };
  }
}
