export default function sseListener(url: string, callback?: (event: string, data: any) => void) {
  let isSubscribed = true;

  const runFetch = async () => {
    return fetch(url, { headers: { Accept: 'text/event-stream' } });
  };

  const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

  const listener = async () => {
    while (isSubscribed) {
      let resp;
      try {
        resp = await runFetch();
        if (!resp.ok) {
          throw new Error(`HTTP error! status: ${resp.status}`);
        }
      } catch (e) {
        console.error('Connection error:', e);
        console.log('Reconnecting in 5 sec...');
        await delay(5000);
        continue;
      }

      try {
        const reader = resp.body.getReader();
        while (isSubscribed) {
          const { done, value } = await reader.read();
          if (done) {
            return;
          }
          const text = new TextDecoder().decode(value);
          const event = text.split('\n')[0].split(':')[1].trim();
          const data = text
            .split('\n')[1]
            .replace(/data: /, '')
            .trim();
          if (callback) {
            let parsed;
            try {
              parsed = JSON.parse(data);
            } catch (e) {
              parsed = data;
            } finally {
              callback(event, parsed);
            }
          }
        }
      } catch (e) {
        console.error('Error reading stream:', e);
        console.log('Reconnecting in 5 sec...');
        await delay(5000); // Wait for 5 seconds before retrying
      }
    }
  };

  listener();

  return () => {
    isSubscribed = false;
  };
}
