function supportsDeclarativeShadowDOM() {
  return HTMLTemplateElement.prototype.hasOwnProperty('shadowRootMode');
}
(function attachShadowRoots(root: Document | ShadowRoot) {
  if (supportsDeclarativeShadowDOM()) {
    // Declarative Shadow DOM is supported, no need to polyfill.
    return;
  }
  root.querySelectorAll<HTMLTemplateElement>('template[shadowrootmode]').forEach((template: HTMLTemplateElement) => {
    const mode = template.getAttribute('shadowrootmode') as 'closed' | 'open';
    const shadowRoot = (template.parentNode as HTMLElement).attachShadow({ mode });

    shadowRoot.appendChild(template.content);
    template.remove();
    attachShadowRoots(shadowRoot);
  });
})(document);
