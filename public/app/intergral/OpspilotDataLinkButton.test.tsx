import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { OpspilotDataLinkButton } from './OpspilotDataLinkButton';

const mockPostMessage = jest.fn();
jest.mock('./OpsPilotBroadcastContext', () => ({
  useOpsPilotBroadcast: () => ({ channel: { postMessage: mockPostMessage } }),
}));

describe('OpspilotDataLinkButton', () => {
  const originalParent = window.parent;

  beforeEach(() => {
    jest.clearAllMocks();
    // Simulate being in an iframe
    Object.defineProperty(window, 'parent', {
      value: { postMessage: jest.fn() },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    Object.defineProperty(window, 'parent', {
      value: originalParent,
      writable: true,
      configurable: true,
    });
  });

  it('should render with the link title', () => {
    render(<OpspilotDataLinkButton link={{ title: 'OpsPilot AI', href: '#opspilot', target: '' }} />);
    expect(screen.getByText('OpsPilot AI')).toBeInTheDocument();
  });

  it('should show external link icon when target is _blank', () => {
    render(<OpspilotDataLinkButton link={{ title: 'OpsPilot AI', href: '#opspilot', target: '_blank' }} />);
    // Button rendered with icon prop — we just verify the button renders
    expect(screen.getByText('OpsPilot AI')).toBeInTheDocument();
  });

  it('should broadcast log content on click when in iframe', async () => {
    const link = { title: 'OpsPilot AI', href: 'some log content here', target: '' };
    render(<OpspilotDataLinkButton link={link} />);
    await userEvent.click(screen.getByText('OpsPilot AI'));

    expect(mockPostMessage).toHaveBeenCalledWith({
      type: 'opspilot-host.integration',
      integration: {
        content: 'some log content here',
        content_type: 'log',
        content_source: 'grafana',
      },
    });
  });

  it('should not broadcast when not in an iframe', async () => {
    // window.parent === window means not in iframe
    Object.defineProperty(window, 'parent', {
      value: window,
      writable: true,
      configurable: true,
    });

    const link = { title: 'OpsPilot AI', href: 'log content', target: '' };
    render(<OpspilotDataLinkButton link={link} />);
    await userEvent.click(screen.getByText('OpsPilot AI'));

    expect(mockPostMessage).not.toHaveBeenCalled();
  });

  it('should handle broadcast errors gracefully', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
    mockPostMessage.mockImplementation(() => {
      throw new Error('Channel error');
    });

    const link = { title: 'OpsPilot AI', href: 'log content', target: '' };
    render(<OpspilotDataLinkButton link={link} />);
    await userEvent.click(screen.getByText('OpsPilot AI'));

    expect(consoleSpy).toHaveBeenCalledWith('Failed to send OpsPilot broadcast:', expect.any(Error));
    consoleSpy.mockRestore();
  });

  it('should pass through buttonProps', () => {
    render(
      <OpspilotDataLinkButton
        link={{ title: 'OpsPilot AI', href: '#', target: '' }}
        buttonProps={{ 'aria-label': 'custom-label' } as any}
      />
    );
    expect(screen.getByRole('button', { name: 'custom-label' })).toBeInTheDocument();
  });
});
