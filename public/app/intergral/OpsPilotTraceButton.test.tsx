import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { Trace } from '../features/explore/TraceView/components/types/trace';

import OpsPilotTraceButton from './OpsPilotTraceButton';

const mockPostMessage = jest.fn();
jest.mock('./OpsPilotBroadcastContext', () => ({
  useOpsPilotBroadcast: () => ({ channel: { postMessage: mockPostMessage } }),
}));

jest.mock('./opspilotStringify', () => ({
  opsPilotStringify: (obj: unknown) => JSON.stringify(obj),
}));

jest.mock('../features/explore/TraceView/components/TracePageHeader/Actions/ActionButton', () => {
  return {
    __esModule: true,
    default: (props: { onClick: () => void; label: string; ariaLabel: string }) => (
      <button onClick={props.onClick} aria-label={props.ariaLabel}>
        {props.label}
      </button>
    ),
  };
});

function makeTrace(overrides: Partial<Trace> = {}): Trace {
  return {
    traceID: 'trace-1',
    traceName: 'test-trace',
    processes: { p1: { serviceName: 'svc', tags: [] } },
    spans: [
      {
        spanID: 'span-1',
        traceID: 'trace-1',
        processID: 'p1',
        operationName: 'op',
        startTime: 1000,
        duration: 500,
        logs: [],
        tags: [],
        references: [{ refType: 'CHILD_OF', spanID: 's0', traceID: 'trace-1' }],
        warnings: [],
        flags: 0,
        childSpanIds: [],
        depth: 0,
        hasChildren: false,
        childSpanCount: 0,
        process: { serviceName: 'svc', tags: [] },
        relativeStartTime: 0,
        subsidiarilyReferencedBy: [],
      },
    ],
    duration: 500,
    startTime: 1000,
    endTime: 1500,
    services: [],
    ...overrides,
  } as unknown as Trace;
}

describe('OpsPilotTraceButton', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render nothing when trace is undefined', () => {
    const { container } = render(<OpsPilotTraceButton trace={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it('should render the Analyze Trace button when trace is provided', () => {
    render(<OpsPilotTraceButton trace={makeTrace()} />);
    expect(screen.getByText('Analyze Trace')).toBeInTheDocument();
  });

  it('should broadcast trace content on click', async () => {
    render(<OpsPilotTraceButton trace={makeTrace()} />);
    await userEvent.click(screen.getByText('Analyze Trace'));

    expect(mockPostMessage).toHaveBeenCalledWith({
      type: 'opspilot-host.integration',
      integration: expect.objectContaining({
        content_type: 'trace',
        content_source: 'grafana',
      }),
    });
  });

  it('should clear span references before broadcasting', async () => {
    const trace = makeTrace();
    expect(trace.spans[0].references).toHaveLength(1);

    render(<OpsPilotTraceButton trace={trace} />);
    await userEvent.click(screen.getByText('Analyze Trace'));

    // After click, references should have been cleared
    expect(trace.spans[0].references).toEqual([]);
  });

  it('should handle postMessage errors gracefully', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
    mockPostMessage.mockImplementation(() => {
      throw new Error('Channel closed');
    });

    render(<OpsPilotTraceButton trace={makeTrace()} />);
    await userEvent.click(screen.getByText('Analyze Trace'));

    expect(consoleSpy).toHaveBeenCalledWith('Failed to analyze trace:', expect.any(Error));
    consoleSpy.mockRestore();
  });
});
