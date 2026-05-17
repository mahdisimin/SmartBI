const LOCAL_STORAGE_KEY = 'smartbi.apiBaseUrl';

export function getInitialBaseUrl() {
  return localStorage.getItem(LOCAL_STORAGE_KEY) || window.SMARTBI_API_BASE_URL || '';
}

export function saveBaseUrl(baseUrl) {
  localStorage.setItem(LOCAL_STORAGE_KEY, baseUrl.trim());
}

export function createApiClient(getBaseUrl) {
  return {
    async request(endpoint, payload) {
      const url = buildUrl(getBaseUrl(), endpoint.path);
      const options = {
        method: endpoint.method,
        headers: {},
      };

      if (endpoint.method !== 'GET') {
        options.headers['Content-Type'] = 'application/json';
        options.body = JSON.stringify(payload);
      }

      const response = await fetch(url, options);
      const responseBody = await parseResponse(response);
      const result = {
        ok: response.ok,
        status: response.status,
        statusText: response.statusText,
        data: responseBody,
      };

      if (!response.ok) {
        const error = new Error(`Request failed with status ${response.status}`);
        error.result = result;
        throw error;
      }

      return result;
    },
  };
}

function buildUrl(baseUrl, path) {
  const normalizedBaseUrl = baseUrl.trim().replace(/\/$/, '');
  return `${normalizedBaseUrl}${path}`;
}

async function parseResponse(response) {
  const text = await response.text();

  if (!text) {
    return null;
  }

  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}
