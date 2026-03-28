import { renderHook } from '@testing-library/react';

import { locationService } from '@grafana/runtime';

import { useIframeNavigation } from './useIframeNavigation';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  locationService: {
    push: jest.fn(),
  },
}));

describe('useIframeNavigation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should navigate when a valid navigate message is received', () => {
    renderHook(() => useIframeNavigation());

    window.dispatchEvent(
      new MessageEvent('message', {
        data: { type: 'navigate', path: '/d/my-dashboard' },
      })
    );

    expect(locationService.push).toHaveBeenCalledWith('/d/my-dashboard');
  });

  it('should ignore messages without navigate type', () => {
    renderHook(() => useIframeNavigation());

    window.dispatchEvent(
      new MessageEvent('message', {
        data: { type: 'other-event', path: '/d/my-dashboard' },
      })
    );

    expect(locationService.push).not.toHaveBeenCalled();
  });

  it('should ignore navigate messages with non-string path', () => {
    renderHook(() => useIframeNavigation());

    window.dispatchEvent(
      new MessageEvent('message', {
        data: { type: 'navigate', path: 123 },
      })
    );

    expect(locationService.push).not.toHaveBeenCalled();
  });

  it('should ignore messages without data', () => {
    renderHook(() => useIframeNavigation());

    window.dispatchEvent(new MessageEvent('message', { data: null }));

    expect(locationService.push).not.toHaveBeenCalled();
  });

  it('should clean up event listener on unmount', () => {
    const { unmount } = renderHook(() => useIframeNavigation());
    unmount();

    window.dispatchEvent(
      new MessageEvent('message', {
        data: { type: 'navigate', path: '/d/test' },
      })
    );

    expect(locationService.push).not.toHaveBeenCalled();
  });
});
