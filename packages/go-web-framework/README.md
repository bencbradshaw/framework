# GoWebFramework frontend

## use with [framework](https://github.com/bencbradshaw/framework)

## quickstart

```bash
npm install go-web-framework
```

autoreload on save:

```javascript
import sse from 'go-web-framework/sse.js';
sse('/events');
// you will now get:
// - auto-reload on changes to js/css files
```

frontend router:

```javascript
import { Router } from 'go-web-framework/router.js';
import { LitElement } from 'lit';
import { customElement } from 'lit/decorators.js';

@customElement('app-root')
export class AppRoot extends LitElement {
  connectedCallback(): void {
    super.connectedCallback();
    const router = new Router(this);
    router.baseUrl = '/app';
    router.addRoute({
      path: '/',
      component: 'chat-route',
      importer: () => import('./routes/chat-route.js'),
      title: 'Chat'
    });
    router.navigate(window.location.pathname);
  }
}
```
