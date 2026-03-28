import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { TraceSpan } from '../features/explore/TraceView/components/types/trace';

import OpsPilotSpanButton from './OpsPilotSpanButton';

const mockPostMessage = jest.fn();
jest.mock('./OpsPilotBroadcastContext', () => ({
  useOpsPilotBroadcast: () => ({ channel: { postMessage: mockPostMessage } }),
}));

jest.mock('./opspilotStringify', () => ({
  opsPilotStringify: (obj: unknown) => JSON.stringify(obj),
}));

function makeSpan(overrides: Partial<TraceSpan> = {}): TraceSpan {
  return {
    spanID: 'span-1',
    traceID: 'trace-1',
    processID: 'p1',
    operationName: 'GET /api',
    startTime: 1000,
    duration: 500,
    logs: [],
    tags: [],
    kind: '',
    statusCode: 0,
    statusMessage: '',
    references: [],
    warnings: [],
    flags: 0,
    childSpanIds: [],
    depth: 0,
    hasChildren: false,
    childSpanCount: 0,
    process: { serviceName: 'test-svc', tags: [] },
    relativeStartTime: 0,
    subsidiarilyReferencedBy: [],
    ...overrides,
  } as TraceSpan;
}

describe('OpsPilotSpanButton', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render the Ask OpsPilot button', () => {
    render(<OpsPilotSpanButton span={makeSpan()} />);
    expect(screen.getByText('Ask OpsPilot')).toBeInTheDocument();
  });

  it('should show Analyze Span menu item on click', async () => {
    render(<OpsPilotSpanButton span={makeSpan()} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    expect(screen.getByText('Analyze Span')).toBeInTheDocument();
  });

  it('should NOT show Analyze Error when statusCode is not 2', async () => {
    render(<OpsPilotSpanButton span={makeSpan({ statusCode: 0 })} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    expect(screen.queryByText('Analyze Error')).not.toBeInTheDocument();
  });

  it('should show Analyze Error when statusCode=2 and reason tag present', async () => {
    const span = makeSpan({
      statusCode: 2,
      tags: [{ key: 'reason', value: 'NullPointerException' }],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    expect(screen.getByText('Analyze Error')).toBeInTheDocument();
  });

  it('should NOT show Analyze Error when statusCode=2 but no reason tag', async () => {
    const span = makeSpan({
      statusCode: 2,
      tags: [{ key: 'other', value: 'val' }],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    expect(screen.queryByText('Analyze Error')).not.toBeInTheDocument();
  });

  it('should show Analyze Query when flv=JDBCRequest', async () => {
    const span = makeSpan({
      tags: [{ key: 'flv', value: 'JDBCRequest' }],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    expect(screen.getByText('Analyze Query')).toBeInTheDocument();
  });

  it('should NOT show Analyze Query when flv is not JDBCRequest', async () => {
    const span = makeSpan({
      tags: [{ key: 'flv', value: 'HTTPRequest' }],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    expect(screen.queryByText('Analyze Query')).not.toBeInTheDocument();
  });

  it('should broadcast span content when Analyze Span is clicked', async () => {
    const span = makeSpan();
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    await userEvent.click(screen.getByText('Analyze Span'));

    expect(mockPostMessage).toHaveBeenCalledWith({
      type: 'opspilot-host.integration',
      integration: expect.objectContaining({
        content_type: 'span',
        content_source: 'grafana',
      }),
    });
  });

  it('should broadcast error content when Analyze Error is clicked', async () => {
    const span = makeSpan({
      statusCode: 2,
      tags: [
        { key: 'reason', value: 'NullPointerException' },
        { key: 'throType', value: 'java.lang.NullPointerException' },
        { key: 'flv', value: 'WebRequest' },
        { key: 'unrelated', value: 'should-be-excluded' },
      ],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    await userEvent.click(screen.getByText('Analyze Error'));

    expect(mockPostMessage).toHaveBeenCalledWith({
      type: 'opspilot-host.integration',
      integration: expect.objectContaining({
        content_type: 'error',
        content_source: 'grafana',
      }),
    });
  });

  it('should broadcast jdbc content when Analyze Query is clicked', async () => {
    const span = makeSpan({
      tags: [
        { key: 'flv', value: 'JDBCRequest' },
        { key: 'db.statement', value: 'SELECT * FROM users' },
        { key: 'db.type', value: 'postgresql' },
      ],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    await userEvent.click(screen.getByText('Analyze Query'));

    expect(mockPostMessage).toHaveBeenCalledWith({
      type: 'opspilot-host.integration',
      integration: expect.objectContaining({
        content_type: 'jdbc',
        content_source: 'grafana',
      }),
    });
  });

  it('should clear references before broadcasting', async () => {
    const span = makeSpan({
      references: [{ refType: 'CHILD_OF', spanID: 's2', traceID: 't1' }],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    await userEvent.click(screen.getByText('Analyze Span'));

    // The span passed to stringify should have references cleared
    const call = mockPostMessage.mock.calls[0][0];
    const content = JSON.parse(call.integration.content);
    expect(content.references).toEqual([]);
  });

  it('should filter generic tags to flv, app_name, tann for error analysis', async () => {
    const span = makeSpan({
      statusCode: 2,
      tags: [
        { key: 'flv', value: 'WebRequest' },
        { key: 'app_name', value: 'myapp' },
        { key: 'tann', value: 'some-annotation' },
        { key: 'reason', value: 'Error occurred' },
        { key: 'unrelated_tag', value: 'should-be-excluded' },
      ],
    });
    render(<OpsPilotSpanButton span={span} />);
    await userEvent.click(screen.getByText('Ask OpsPilot'));
    await userEvent.click(screen.getByText('Analyze Error'));

    const call = mockPostMessage.mock.calls[0][0];
    const content = JSON.parse(call.integration.content);
    // tags should include flv, app_name, tann (generic) + reason (error-specific)
    const tagKeys = content.tags.map((t: { key: string }) => t.key);
    expect(tagKeys).toContain('flv');
    expect(tagKeys).toContain('app_name');
    expect(tagKeys).toContain('tann');
    expect(tagKeys).toContain('reason');
    expect(tagKeys).not.toContain('unrelated_tag');
  });
});
