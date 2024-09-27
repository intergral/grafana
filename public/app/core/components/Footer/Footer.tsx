import { css } from '@emotion/css';
import { memo } from 'react';

import { GrafanaTheme2, LinkTarget } from '@grafana/data';
import { IconName, useStyles2 } from '@grafana/ui';
import { t } from 'app/core/internationalization';

export interface FooterLink {
  target: LinkTarget;
  text: string;
  id: string;
  icon?: IconName;
  url?: string;
}

export let getFooterLinks = (): FooterLink[] => {
  return [
    {
      target: '_blank',
      id: 'documentation',
      text: t('nav.help/documentation', 'Documentation'),
      icon: 'document-info',
      url: 'https://grafana.com/docs/grafana/latest/?utm_source=grafana_footer',
    },
    {
      target: '_blank',
      id: 'support',
      text: t('nav.help/support', 'Support'),
      icon: 'question-circle',
      url: 'https://grafana.com/products/enterprise/?utm_source=grafana_footer',
    },
    {
      target: '_blank',
      id: 'community',
      text: t('nav.help/community', 'Community'),
      icon: 'comments-alt',
      url: 'https://community.grafana.com/?utm_source=grafana_footer',
    },
  ];
};


export interface Props {
  /** Link overrides to show specific links in the UI */
  customLinks?: FooterLink[] | null;
  hideEdition?: boolean;
}

export const Footer = memo(({ }: Props) => {
  const styles = useStyles2(getStyles);

  return (
    <footer className={styles.footer}>
    </footer>
  );
});

Footer.displayName = 'Footer';


const getStyles = (theme: GrafanaTheme2) => ({
  footer: css({
    ...theme.typography.bodySmall,
    color: theme.colors.text.primary,
    display: 'block',
    padding: theme.spacing(2, 0),
    position: 'relative',
    width: '98%',

    'a:hover': {
      color: theme.colors.text.maxContrast,
      textDecoration: 'underline',
    },

    [theme.breakpoints.down('md')]: {
      display: 'none',
    },
  }),
  list: css({
    listStyle: 'none',
  }),
  listItem: css({
    display: 'inline-block',
    '&:after': {
      content: "' | '",
      padding: theme.spacing(0, 1),
    },
    '&:last-child:after': {
      content: "''",
      paddingLeft: 0,
    },
  }),
});
