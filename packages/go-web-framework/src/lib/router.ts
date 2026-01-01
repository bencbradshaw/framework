type Route = {
  path: string;
  component?: string;
  importer?: () => Promise<any>;
  redirect?: string;
  title?: string;
  params?: Record<string, string>;
};

/**
 * A client-side router for single-page applications with support for dynamic route parameters.
 * Handles browser navigation (back/forward), link click interception, and dynamic component loading.
 *
 * Notes:
 * - `baseUrl` is prepended when registering routes via `addRoute()`.
 * - `navigate()` expects a full path (including the baseUrl) and uses `history.pushState()`.
 * - The router intercepts `<a href="...">` clicks by default; add `router-ignore` to opt out.
 * - Call `destroy()` when the router is no longer needed to remove global event listeners.
 *
 * @example
 * ```typescript
 * const router = new Router(document.getElementById('app'));
 * router.baseUrl = '/app';
 *
 * // Add exact match route
 * router.addRoute({
 *   path: '/',
 *   component: 'home-page',
 *   importer: () => import('./pages/home.js'),
 *   title: 'Home'
 * });
 *
 * // Add parameterized route
 * router.addRoute({
 *   path: '/chat/:id',
 *   component: 'chat-page',
 *   importer: () => import('./pages/chat.js'),
 *   title: 'Chat'
 * });
 *
 * // Navigate programmatically
 * router.navigate('/app/chat/123');
 *
 * // Access route params
 * const chatId = router.activeRoute.params?.id; // "123"
 *
 * // Add catch-all route with redirect (add this last)
 * router.addRoute({
 *   path: '*',
 *   redirect: '/app/'
 * });
 *
 * // Or render a not-found component
 * router.addRoute({
 *   path: '*',
 *   component: 'not-found-page',
 *   importer: () => import('./pages/not-found.js'),
 *   title: 'Not Found'
 * });
 * ```
 */
export class Router {
  private routes: Route[] = [];
  private rootElement: HTMLElement;

  /**
   * Base URL prefix for all routes. Set this before adding routes.
   */
  public baseUrl: string;

  /**
   * Currently active route with matched parameters.
   */
  public activeRoute: Route;

  /**
   * Creates a new Router instance.
   * @param rootElement - The HTML element that will contain routed components
   */
  constructor(rootElement: HTMLElement) {
    this.rootElement = rootElement;

    window.addEventListener('popstate', this.handlePopState);
    document.addEventListener('click', this.handleLinkClick);
  }

  /**
   * Removes global event listeners added by the router.
   * Call this to avoid leaking handlers when creating/destroying routers.
   */
  public destroy(): void {
    window.removeEventListener('popstate', this.handlePopState);
    document.removeEventListener('click', this.handleLinkClick);
  }

  /**
   * Registers a new route. Supports dynamic parameters using `:param` syntax.
   * Routes can either render a component or redirect to another path.
   * @param route - Route configuration object
   * @example
   * ```typescript
   * // Render a component
   * router.addRoute({
   *   path: '/user/:id/post/:postId',
   *   component: 'post-view',
   *   importer: () => import('./post.js'),
   *   title: 'Post'
   * });
   *
   * // Redirect to another path
   * router.addRoute({
   *   path: '/old-path',
   *   redirect: '/app/new-path'
   * });
   * ```
   */
  public addRoute(route: Route) {
    const fullPath = this.baseUrl + route.path;
    this.routes.push({ ...route, path: fullPath });
  }

  /**
   * Programmatically navigates to a new path and updates browser history.
   * @param path - Full path to navigate to (including baseUrl)
   * @example
   * ```typescript
   * router.navigate('/app/chat/123');
   * ```
   */
  public navigate(path: string) {
    history.pushState({}, '', path);
    this.handleRouteChange(path);
  }

  private handlePopState = (): void => {
    this.handleRouteChange(window.location.pathname);
  };

  private handleLinkClick = (event: MouseEvent): void => {
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
  };

  private matchRoute(pathname: string): { route: Route; params: Record<string, string> } | null {
    for (const route of this.routes) {
      // Handle wildcard catch-all
      const pathWithoutBase = route.path.replace(this.baseUrl, '');
      if (pathWithoutBase === '*' || pathWithoutBase === '/*' || route.path === '*' || route.path.endsWith('/*')) {
        return { route, params: {} };
      }

      const pattern = route.path.replace(/:[^/]+/g, '([^/]+)');
      const regex = new RegExp(`^${pattern}$`);
      const match = pathname.match(regex);

      if (match) {
        const paramNames = (route.path.match(/:[^/]+/g) || []).map((p) => p.slice(1));
        const params: Record<string, string> = {};
        paramNames.forEach((name, i) => {
          params[name] = match[i + 1];
        });
        return { route, params };
      }
    }
    return null;
  }

  private handleRouteChange(path: string) {
    const url = new URL(path, window.location.origin);
    const pathname = url.pathname;

    const match = this.matchRoute(pathname);

    if (match) {
      const { route, params } = match;

      // Handle redirect
      if (route.redirect) {
        this.navigate(route.redirect);
        return;
      }

      // Render component
      if (route.importer && route.component) {
        route
          .importer()
          .then(() => {
            while (this.rootElement?.shadowRoot?.firstChild) {
              this.rootElement.shadowRoot.removeChild(this.rootElement.shadowRoot.firstChild);
            }
            const theInnerElement = document.createElement(route.component!);
            this.rootElement.shadowRoot.appendChild(theInnerElement);
            document.title = route.title ?? route.component!;
            this.activeRoute = { ...route, params };
          })
          .catch((error) => {
            console.error('Error importing module:', error);
          });
      }
    } else {
      console.warn('No route found for path:', pathname);
    }
  }
}
