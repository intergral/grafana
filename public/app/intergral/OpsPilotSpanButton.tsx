import { TraceKeyValuePair } from '@grafana/data';
import { Button, Dropdown, Menu } from '@grafana/ui';

import { TraceSpan } from '../features/explore/TraceView/components/types/trace';

import { useOpsPilotBroadcast } from './OpsPilotBroadcastContext';
import { opsPilotStringify } from './opspilotStringify';

export interface OpsPilotSpanButtonProps {
  span: TraceSpan;
}

export default function OpsPilotSpanButton({ span }: OpsPilotSpanButtonProps) {
  const { channel } = useOpsPilotBroadcast();

  const getGenericTags = (customTags: TraceKeyValuePair[]) => {
    return {
      tags: span.tags
        .filter((tag) => tag.key === 'flv' || tag.key === 'app_name' || tag.key === 'tann')
        .concat(customTags),
      process: span.process.tags,
      serviceName: span.process.serviceName,
      logs: span.logs,
      startTime: span.startTime,
      duration: span.duration,
      spanID: span.spanID,
      operationName: span.operationName,
      traceID: span.traceID,
      childSpanIds: span.childSpanIds,
      statusMessage: span.statusMessage,
      ...('parentSpanID' in span && span.parentSpanID ? { parentSpanId: span.parentSpanID } : {}),
    };
  };

  const sendBroadcast = (content: object, contentType: string) => {
    try {
      const broadcastIntegration = {
        content: opsPilotStringify(content),
        content_type: contentType,
        content_source: 'grafana'
      };
      channel?.postMessage({ type: 'opspilot-host.integration', broadcastIntegration });
    } catch (error) {
      console.error('Failed to send OpsPilot broadcast:', error);
    }
  };

  const handleSpanClick = () => {
    try {
      sendBroadcast(span, 'span');
    } catch (error) {
      console.error('Failed to analyze span:', error);
    }
  };

  const handleErrorClick = () => {
    try {
      const errorTags = span.tags.filter(
        (tag) =>
          tag.key === 'reason' || tag.key === 'throType' || tag.key === 'throwable' || tag.key === 'http.status_code'
      );
      sendBroadcast(getGenericTags(errorTags), 'error');
    } catch (error) {
      console.error('Failed to analyze error:', error);
    }
  };

  const handleJdbcClick = () => {
    try {
      const jdbcTags = span.tags.filter((tag) => tag.key.startsWith('db.'));
      sendBroadcast(getGenericTags(jdbcTags), 'jdbc');
    } catch (error) {
      console.error('Failed to analyze query:', error);
    }
  };

  const menu = (
    <Menu>
      <Menu.Item label="Analyze Span" onClick={handleSpanClick} />
      {span.statusCode === 2 && span.tags.find((tag) => tag.key === 'reason') && (
        <Menu.Item label="Analyze Error" onClick={handleErrorClick} />
      )}
      {span.tags.find((tag) => tag.key === 'flv')?.value === 'JDBCRequest' && (
        <Menu.Item label="Analyze Query" onClick={handleJdbcClick} />
      )}
    </Menu>
  );

  return (
    <Dropdown overlay={menu}>
      <Button variant="primary" size="sm" icon="ai">
        Ask OpsPilot
      </Button>
    </Dropdown>
  );
}
