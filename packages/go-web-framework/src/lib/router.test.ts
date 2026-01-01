import type { MockInstance } from 'vitest';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Router } from './router.js';

describe('Router', () => {
  let rootElement: HTMLElement;
  let router: Router;
  let pushStateSpy: MockInstance;

  beforeEach(() => {
    rootElement = document.createElement('div');
    rootElement.attachShadow({ mode: 'open' });
    router = new Router(rootElement);
    router.baseUrl = '/app';

    // In Vitest browser mode, changing the iframe URL (pushState or default <a> navigation)
    // breaks the orchestrator connection. Neutralize pushState so Router.navigate can't
    // accidentally move the iframe away from the session URL.
    pushStateSpy = vi.spyOn(window.history, 'pushState').mockImplementation(() => {});
  });

  afterEach(() => {
    router.destroy();

    // Clean up any links that might have been added
    document.querySelectorAll('a').forEach((link) => link.remove());
    vi.restoreAllMocks();
  });

  it('accepts and registers routes', () => {
    router.addRoute({
      path: '/',
      component: 'home-page',
      importer: () => Promise.resolve(),
      title: 'Home'
    });

    router.addRoute({
      path: '/chat/:id',
      component: 'chat-page',
      importer: () => Promise.resolve(),
      title: 'Chat'
    });

    expect(router).toBeDefined();
  });

  it('navigates to the specified route when URL matches', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/',
      component: 'home-page',
      importer: mockImporter,
      title: 'Home'
    });

    router.navigate('/app/');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockImporter).toHaveBeenCalled();
    expect(rootElement.shadowRoot?.querySelector('home-page')).toBeDefined();
    expect(document.title).toBe('Home');
    expect(router.activeRoute.path).toBe('/app/');
  });

  it('matches parameterized routes and extracts params', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/chat/:id',
      component: 'chat-page',
      importer: mockImporter,
      title: 'Chat'
    });

    router.navigate('/app/chat/123');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockImporter).toHaveBeenCalled();
    expect(router.activeRoute.params?.id).toBe('123');
    expect(rootElement.shadowRoot?.querySelector('chat-page')).toBeDefined();
  });

  it('handles catch-all routes', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/',
      component: 'home-page',
      importer: () => Promise.resolve(),
      title: 'Home'
    });

    router.addRoute({
      path: '*',
      component: 'not-found-page',
      importer: mockImporter,
      title: 'Not Found'
    });

    router.navigate('/app/unknown-path');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockImporter).toHaveBeenCalled();
    expect(rootElement.shadowRoot?.querySelector('not-found-page')).toBeDefined();
  });

  it('handles redirect routes', () => {
    const navigateSpy = vi.spyOn(router, 'navigate');

    router.addRoute({
      path: '/',
      component: 'home-page',
      importer: () => Promise.resolve(),
      title: 'Home'
    });

    router.addRoute({
      path: '/old-path',
      redirect: '/app/'
    });

    router.navigate('/app/old-path');

    expect(navigateSpy).toHaveBeenCalledWith('/app/');
  });

  it('matches more specific routes over less specific ones', async () => {
    const genericMockImporter = vi.fn(() => Promise.resolve());
    const specificMockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/chat/:id',
      component: 'chat-page',
      importer: genericMockImporter,
      title: 'Chat'
    });

    router.addRoute({
      path: '/chat/:id/something',
      component: 'chat-something-page',
      importer: specificMockImporter,
      title: 'Chat Something'
    });

    router.navigate('/app/chat/123/something');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(specificMockImporter).toHaveBeenCalled();
    expect(genericMockImporter).not.toHaveBeenCalled();
    expect(rootElement.shadowRoot?.querySelector('chat-something-page')).toBeDefined();
    expect(router.activeRoute.params?.id).toBe('123');
  });

  it('respects route order - first matching route wins', async () => {
    const firstMockImporter = vi.fn(() => Promise.resolve());
    const secondMockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/page/:slug',
      component: 'generic-page',
      importer: firstMockImporter,
      title: 'Generic'
    });

    router.addRoute({
      path: '/page/:slug',
      component: 'duplicate-page',
      importer: secondMockImporter,
      title: 'Duplicate'
    });

    router.navigate('/app/page/test');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(firstMockImporter).toHaveBeenCalled();
    expect(secondMockImporter).not.toHaveBeenCalled();
    expect(rootElement.shadowRoot?.querySelector('generic-page')).toBeDefined();
  });

  it('handles multiple parameters in a single route', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/user/:userId/post/:postId',
      component: 'post-page',
      importer: mockImporter,
      title: 'Post'
    });

    router.navigate('/app/user/42/post/99');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockImporter).toHaveBeenCalled();
    expect(router.activeRoute.params?.userId).toBe('42');
    expect(router.activeRoute.params?.postId).toBe('99');
  });

  it('handles trailing slashes correctly', async () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/about',
      component: 'about-page',
      importer: mockImporter,
      title: 'About'
    });

    router.navigate('/app/about/');

    await new Promise((resolve) => setTimeout(resolve, 0));

    // Should not match due to trailing slash difference
    expect(mockImporter).not.toHaveBeenCalled();

    warnSpy.mockRestore();
  });

  it('handles navigation to same route', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/page',
      component: 'test-page',
      importer: mockImporter,
      title: 'Test'
    });

    router.navigate('/app/page');
    await new Promise((resolve) => setTimeout(resolve, 0));

    router.navigate('/app/page');
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Should call importer twice
    expect(mockImporter).toHaveBeenCalledTimes(2);
  });

  it('handles importer errors gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const mockImporter = vi.fn(() => Promise.reject(new Error('Import failed')));

    router.addRoute({
      path: '/error',
      component: 'error-page',
      importer: mockImporter,
      title: 'Error'
    });

    router.navigate('/app/error');

    await new Promise((resolve) => setTimeout(resolve, 10));

    expect(mockImporter).toHaveBeenCalled();
    expect(consoleSpy).toHaveBeenCalledWith('Error importing module:', expect.any(Error));

    consoleSpy.mockRestore();
  });

  it('handles redirect chains', () => {
    const navigateSpy = vi.spyOn(router, 'navigate');

    router.addRoute({
      path: '/old',
      redirect: '/app/intermediate'
    });

    router.addRoute({
      path: '/intermediate',
      redirect: '/app/final'
    });

    router.addRoute({
      path: '/final',
      component: 'final-page',
      importer: () => Promise.resolve(),
      title: 'Final'
    });

    router.navigate('/app/old');

    // Should follow redirect chain
    expect(navigateSpy).toHaveBeenCalledWith('/app/intermediate');
  });

  it('handles browser back/forward navigation', async () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/page1',
      component: 'page-one',
      importer: () => Promise.resolve(),
      title: 'Page 1'
    });

    router.addRoute({
      path: '/page2',
      component: 'page-two',
      importer: mockImporter,
      title: 'Page 2'
    });

    router.navigate('/app/page1');
    router.navigate('/app/page2');

    // Simulate back button (popstate event)
    window.dispatchEvent(new PopStateEvent('popstate'));

    await new Promise((resolve) => setTimeout(resolve, 0));

    // Should re-evaluate current location
    expect(mockImporter).toHaveBeenCalled();

    warnSpy.mockRestore();
  });

  it('matches exact paths over parameterized paths when defined first', async () => {
    const exactMockImporter = vi.fn(() => Promise.resolve());
    const paramMockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/users/new',
      component: 'new-user-page',
      importer: exactMockImporter,
      title: 'New User'
    });

    router.addRoute({
      path: '/users/:id',
      component: 'user-page',
      importer: paramMockImporter,
      title: 'User'
    });

    router.navigate('/app/users/new');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(exactMockImporter).toHaveBeenCalled();
    expect(paramMockImporter).not.toHaveBeenCalled();
  });

  it('preserves query parameters in URL', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/search',
      component: 'search-page',
      importer: mockImporter,
      title: 'Search'
    });

    pushStateSpy.mockClear();

    router.navigate('/app/search?q=test&page=2');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockImporter).toHaveBeenCalled();
    const lastCall = pushStateSpy.mock.calls.at(-1);
    const capturedUrl = String(lastCall?.[2] ?? '');
    expect(capturedUrl).toContain('q=test');
    expect(capturedUrl).toContain('page=2');
  });

  it('handles hash fragments in URLs', async () => {
    const mockImporter = vi.fn(() => Promise.resolve());

    router.addRoute({
      path: '/docs',
      component: 'docs-page',
      importer: mockImporter,
      title: 'Docs'
    });

    pushStateSpy.mockClear();

    router.navigate('/app/docs#section-2');

    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockImporter).toHaveBeenCalled();
    const lastCall = pushStateSpy.mock.calls.at(-1);
    const capturedUrl = String(lastCall?.[2] ?? '');
    expect(capturedUrl).toContain('#section-2');
  });

  it('intercepts link clicks and navigates without page reload', () => {
    const navigateSpy = vi.spyOn(router, 'navigate').mockImplementation(() => {});

    const link = document.createElement('a');
    link.setAttribute('href', '/app/test');

    const preventDefault = vi.fn();
    const event = {
      composedPath: () => [link],
      preventDefault
    } as unknown as MouseEvent;

    (router as unknown as { handleLinkClick: (e: MouseEvent) => void }).handleLinkClick(event);

    expect(preventDefault).toHaveBeenCalled();
    expect(navigateSpy).toHaveBeenCalledWith('/app/test');
  });

  it('does not intercept links with router-ignore attribute', () => {
    const navigateSpy = vi.spyOn(router, 'navigate');

    const link = document.createElement('a');
    link.setAttribute('href', '/app/test');
    link.setAttribute('router-ignore', '');

    const preventDefault = vi.fn();
    const event = {
      composedPath: () => [link],
      preventDefault
    } as unknown as MouseEvent;

    (router as unknown as { handleLinkClick: (e: MouseEvent) => void }).handleLinkClick(event);

    expect(preventDefault).not.toHaveBeenCalled();
    expect(navigateSpy).not.toHaveBeenCalled();
  });
});
