import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Router } from 'react-router-dom';
import { TestProvider } from 'test/helpers/TestProvider';
import { getGrafanaContextMock } from 'test/mocks/getGrafanaContextMock';

import { NavModelItem } from '@grafana/data';
import { selectors } from '@grafana/e2e-selectors';
import { locationService } from '@grafana/runtime';

import { MegaMenu } from './MegaMenu';

const setup = () => {
  const navBarTree: NavModelItem[] = [
    {
      text: 'Section name',
      id: 'section',
      url: 'section',
      children: [
        {
          text: 'Child1',
          id: 'child1',
          url: 'section/child1',
          children: [{ text: 'Grandchild1', id: 'grandchild1', url: 'section/child1/grandchild1' }],
        },
        { text: 'Child2', id: 'child2', url: 'section/child2' },
      ],
    },
    {
      text: 'Profile',
      id: 'profile',
      url: 'profile',
    },
  ];

  const grafanaContext = getGrafanaContextMock();
  grafanaContext.chrome.setMegaMenuOpen(true);

  return render(
    <TestProvider storeState={{ navBarTree }} grafanaContext={grafanaContext}>
      <Router history={locationService.getHistory()}>
        <MegaMenu onClose={() => {}} />
      </Router>
    </TestProvider>
  );
};

describe('MegaMenu', () => {
  afterEach(() => {
    window.localStorage.clear();
  });
  it('should not render component', async () => {
    setup();

    expect(await screen.queryByTestId(selectors.components.NavMenu.Menu)).not.toBeInTheDocument();
    expect(await screen.queryByRole('link', { name: 'Section name' })).not.toBeInTheDocument();
  });

  it('should filter out profile', async () => {
    setup();

    expect(screen.queryByLabelText('Profile')).not.toBeInTheDocument();
  });
});
