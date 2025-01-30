export default async function sseListener(url: string, callback?: (event: string, data: any) => void) {
  const runFetch = async () => {
    console.log(`Connecting to ${url}...`);
    return fetch(url, { headers: { Accept: 'text/event-stream' } });
  };

  const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

  while (true) {
    let resp;
    try {
      resp = await runFetch();
      console.log(`${url} response:`, resp);
      if (!resp.ok) {
        throw new Error(`HTTP error! status: ${resp.status}`);
      }
    } catch (e) {
      console.error('Connection error:', e);
      console.log('Reconnecting in 5 sec...');
      await delay(5000); // Wait for 5 seconds before retrying
      continue;
    }

    try {
      const reader = resp.body.getReader();
      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          // Do something with last chunk of data then exit reader
          return;
        }
        const text = new TextDecoder().decode(value);
        // text is in form of "event: string\ndata: string\n\n"
        const event = text.split('\n')[0].split(':')[1].trim();
        const data = text
          .split('\n')[1]
          .replace(/data: /, '')
          .trim();
        console.log('event:', event, 'data:', data);
        if (event === 'reload') {
          window.location.reload();
        }
        if (event === 'entity' && callback) {
          console.log('calling callback');
          callback(event, JSON.parse(data));
        }
      }
    } catch (e) {
      console.error('Error reading stream:', e);
      console.log('Reconnecting in 5 sec...');
      await delay(5000); // Wait for 5 seconds before retrying
    }
  }
}
