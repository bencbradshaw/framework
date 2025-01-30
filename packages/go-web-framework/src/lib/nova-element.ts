export class NovaElement extends HTMLElement {
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
