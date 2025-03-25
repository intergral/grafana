import { useEffect } from 'react';

import { getTimeSrv } from 'app/features/dashboard/services/TimeSrv';


export const useOpspilotMetadata = () => {
  useEffect(() => {
    const event = async (event: MessageEvent) => {
      if (event.data.type === 'opspilot-host.getMetadata') {
        window.parent.postMessage({type: "opspilot-slave.sendMetadata", metadata: {
            slaveUrl: window.location.pathname,
            timeStart: getTimeSrv().timeRange().from.valueOf(),
            timeEnd: getTimeSrv().timeRange().to.valueOf(),
            timezone: getTimeSrv().timeModel?.getTimezone() === 'browser' ? Intl.DateTimeFormat().resolvedOptions().timeZone : getTimeSrv().timeModel?.getTimezone(),
          } }, '*');
      }
    };
    window.addEventListener('message', event);
    return () => {
      window.removeEventListener('message', event);
    }
  }, []);

}
