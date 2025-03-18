export function provide(uniqueName: string) {
  return function (target: any, propertyKey: string) {
    const originalConnectedCallback = target.connectedCallback;
    target.connectedCallback = function () {
      if (originalConnectedCallback) {
        originalConnectedCallback.call(this);
      }
      this.addEventListener('context-request', (event: CustomEvent) => {
        if (event.detail.name === uniqueName) {
          event.detail.callback(this[propertyKey]);
        }
      });
    };
    const originalDisconnectedCallback = target.disconnectedCallback;
    target.disconnectedCallback = function () {
      if (originalDisconnectedCallback) {
        originalDisconnectedCallback.call(this);
      }
      this.removeEventListener('context-request', (event: CustomEvent) => {
        if (event.detail.name === uniqueName) {
          event.detail.callback(this[propertyKey]);
        }
      });
    };
  };
}

export function consume(uniqueName: string) {
  return function (target: any, propertyKey: string) {
    const originalConnectedCallback = target.connectedCallback;
    target.connectedCallback = function () {
      const event = new CustomEvent('context-request', {
        detail: {
          name: uniqueName,
          callback: (value: any) => {
            this[propertyKey] = value;
          }
        },
        bubbles: true,
        composed: true
      });
      this.dispatchEvent(event);
      if (originalConnectedCallback) {
        originalConnectedCallback.call(this);
      }
    };
  };
}
