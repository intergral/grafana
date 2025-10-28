import { ButtonProps, Button } from '@grafana/ui';

import { useOpsPilotBroadcast } from './OpsPilotBroadcastContext';

type DataLinkButtonProps = {
  link: any;
  buttonProps?: ButtonProps;
};

/**
 * @internal
 */
export function OpspilotDataLinkButton({ link, buttonProps }: DataLinkButtonProps) {
  const { channel } = useOpsPilotBroadcast();

  const handleClick = () => {
    try {
      if (window.parent !== window) {
        // const fusionReactorIntegration = { content: link.href, content_type: 'log' };
        // window.parent.postMessage({ type: 'opspilot-slave.sendIntegration', fusionReactorIntegration }, '*');
        const broadcastIntegration = { content: link.href, content_type: 'log', content_source: 'grafana' };
        channel?.postMessage({ type: 'opspilot-host.integration', integration: broadcastIntegration });
      }
    } catch (error) {
      console.error('Failed to send OpsPilot broadcast:', error);
    }
  };

  return (
    <Button
      icon={link.target === '_blank' ? 'external-link-alt' : undefined}
      variant="primary"
      size="sm"
      onClick={handleClick}
      {...buttonProps}
    >
      {link.title}
    </Button>
  );
}
