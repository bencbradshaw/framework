/**
 * a minimal base class for web components.
 * this is experimental and is recommended to use an alternative like LitElement
 * @experimental
 * */
export class FrameworkElement extends HTMLElement {
  isUpdateScheduled = false;

  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
    this.update();
  }

  #r: () => void;
  #rj: () => void;
  updateComplete: Promise<void> = new Promise((r, rj) => {
    this.#r = r;
    this.#rj = rj;
  });

  render?(): string;

  update() {
    if (this.shadowRoot && this.render) {
      this.shadowRoot.innerHTML = this?.render();
    }
    this.#r();
  }
}
