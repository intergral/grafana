import { ButtonProps, Button } from '@grafana/ui';

type DataLinkButtonProps = {
  link: any;
  buttonProps?: ButtonProps;
};

/**
 * @internal
 */
export function OpspilotDataLinkButton({ link, buttonProps }: DataLinkButtonProps) {
  return (
    <Button
      icon={link.target === '_blank' ? 'external-link-alt' : undefined}
      variant="primary"
      size="sm"
      onClick={() => {
        if (window.parent !== window) {
          const integration = {content: link.href, content_type: "log"}
          window.parent.postMessage({type: "opspilot-slave.sendIntegration", integration } , "*")
        }
      }}
      {...buttonProps}
    >
      {link.title}
    </Button>
  );
}
