import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Router } from './router.js';

describe('Router', () => {
  let rootElement: HTMLElement;
  let router: Router;

  beforeEach(() => {
    rootElement = document.createElement('div');
    rootElement.attachShadow({ mode: 'open' });
    router = new Router(rootElement);
    router.baseUrl = '/app';

    // Mock history and location
    vi.stubGlobal('history', {
      pushState: vi.fn()
    });
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
      path: '/old-path',
      redirect: '/app/'
    });

    router.navigate('/app/old-path');

    expect(navigateSpy).toHaveBeenCalledWith('/app/');
  });
});
