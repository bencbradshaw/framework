type Route = {
  path: string;
  component: string;
  importer: () => Promise<any>;
  title?: string;
  params?: Record<string, string>;
};

export class Router {
  private routes: Route[] = [];
  private rootElement: HTMLElement;
  public baseUrl: string;
  public activeRoute: Route;

  constructor(rootElement: HTMLElement) {
    this.rootElement = rootElement;
    window.addEventListener('popstate', this.handlePopState.bind(this));
    document.addEventListener('click', this.handleLinkClick.bind(this));
  }

  public addRoute(route: Route) {
    const fullPath = this.baseUrl + route.path;
    this.routes.push({ ...route, path: fullPath });
  }

  public navigate(path: string) {
    history.pushState({}, '', path);
    this.handleRouteChange(path);
  }

  private handlePopState() {
    this.handleRouteChange(window.location.pathname);
  }

  private handleLinkClick(event: MouseEvent) {
    const composedPath0 = event.composedPath()[0] as HTMLElement;
    const composedPath0Parent = composedPath0.parentElement;
    const firstA =
      composedPath0?.tagName === 'A'
        ? composedPath0
        : composedPath0Parent?.tagName === 'A'
        ? composedPath0Parent
        : null;
    if (firstA && !firstA.hasAttribute('router-ignore')) {
      event.preventDefault();
      const path = firstA.getAttribute('href')!;
      this.navigate(path);
    }
  }

  private handleRouteChange(path: string) {
    const route = this.routes.find((route) => route.path === path);
    if (route) {
      route
        .importer()
        .then((module) => {
          while (this.rootElement?.shadowRoot?.firstChild) {
            this.rootElement.shadowRoot.removeChild(this.rootElement.shadowRoot.firstChild);
          }
          const theInnerElement = document.createElement(route.component);
          this.rootElement.shadowRoot.appendChild(theInnerElement);
          document.title = route.title ?? route.component;
          this.activeRoute = route;
        })
        .catch((error) => {
          console.error('Error importing module:', error);
        });
    } else {
      console.warn('No route found for path:', path);
    }
  }
}
