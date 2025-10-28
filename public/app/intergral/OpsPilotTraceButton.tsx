import ActionButton from '../features/explore/TraceView/components/TracePageHeader/Actions/ActionButton';
import { Trace } from '../features/explore/TraceView/components/types';

import { useOpsPilotBroadcast } from './OpsPilotBroadcastContext';
import { opsPilotStringify } from './opspilotStringify';

export interface OpsPilotTraceButtonProps {
  trace?: Trace;
}

export default function OpsPilotTraceButton({ trace }: OpsPilotTraceButtonProps) {
  const { channel } = useOpsPilotBroadcast();

  if (!trace) {
    return null;
  }

  const handleClick = () => {
    try {
      const broadcastIntegration = {
        content: opsPilotStringify(trace),
        content_type: 'trace',
        content_source: 'grafana',
      };
      channel?.postMessage({ type: 'opspilot-host.integration', integration: broadcastIntegration });
    } catch (error) {
      console.error('Failed to analyze trace:', error);
    }
  };

  return <ActionButton onClick={handleClick} ariaLabel={'Analyze Trace'} label={'Analyze Trace'} icon={'ai'} />;
}
