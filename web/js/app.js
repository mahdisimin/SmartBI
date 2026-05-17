import { createApiClient, getInitialBaseUrl, saveBaseUrl } from './apiClient.js';
import { endpoints } from './endpoints.js';

const baseUrlInput = document.querySelector('#base-url');
const endpointList = document.querySelector('#endpoint-list');
const apiClient = createApiClient(() => baseUrlInput.value);

baseUrlInput.value = getInitialBaseUrl();
baseUrlInput.addEventListener('change', () => saveBaseUrl(baseUrlInput.value));
baseUrlInput.addEventListener('blur', () => saveBaseUrl(baseUrlInput.value));

endpointList.append(...endpoints.map(createEndpointCard));

function createEndpointCard(endpoint) {
  const card = document.createElement('article');
  card.className = 'endpoint-card';

  const fieldsMarkup = endpoint.fields.length
    ? `<div class="fields-grid">${endpoint.fields.map(createFieldMarkup).join('')}</div>`
    : '<p class="endpoint-description">No request parameters are required.</p>';

  card.innerHTML = `
    <div class="endpoint-card__header">
      <div>
        <div class="endpoint-meta">
          <span class="method">${endpoint.method}</span>
          <span class="path">${endpoint.path}</span>
        </div>
        <h3>${endpoint.name}</h3>
        <p class="endpoint-description">${endpoint.description}</p>
      </div>
      <span class="status-pill idle" data-status>Idle</span>
    </div>
    <form class="endpoint-form" data-form>
      ${fieldsMarkup}
      <div class="actions">
        <button type="submit" data-submit>Execute Request</button>
      </div>
    </form>
    <div class="result-grid">
      <section class="result-panel">
        <h4>Request</h4>
        <pre data-request>{}</pre>
      </section>
      <section class="result-panel">
        <h4>Response</h4>
        <pre data-response>Run the request to see the response.</pre>
      </section>
    </div>
  `;

  const form = card.querySelector('[data-form]');
  const submitButton = card.querySelector('[data-submit]');
  const status = card.querySelector('[data-status]');
  const requestPanel = card.querySelector('[data-request]');
  const responsePanel = card.querySelector('[data-response]');

  form.addEventListener('submit', async (event) => {
    event.preventDefault();

    const payload = readFormPayload(form, endpoint.fields);
    const requestPreview = {
      method: endpoint.method,
      url: `${baseUrlInput.value.trim()}${endpoint.path}`,
      body: endpoint.method === 'GET' ? undefined : payload,
    };

    requestPanel.textContent = formatJson(requestPreview);
    setStatus(status, 'loading', 'Loading...');
    submitButton.disabled = true;
    responsePanel.textContent = 'Waiting for response...';

    try {
      const result = await apiClient.request(endpoint, payload);
      setStatus(status, 'success', `Success ${result.status}`);
      responsePanel.textContent = formatJson(result);
    } catch (error) {
      const result = error.result || { message: error.message };
      setStatus(status, 'error', 'Error');
      responsePanel.textContent = formatJson(result);
    } finally {
      submitButton.disabled = false;
    }
  });

  return card;
}

function createFieldMarkup(field) {
  const required = field.required ? 'required' : '';
  return `
    <label class="field">
      <span>${field.label}</span>
      <input
        name="${field.name}"
        type="${field.type || 'text'}"
        placeholder="${field.placeholder || ''}"
        autocomplete="off"
        ${required}
      />
    </label>
  `;
}

function readFormPayload(form, fields) {
  return fields.reduce((payload, field) => {
    payload[field.name] = new FormData(form).get(field.name) || '';
    return payload;
  }, {});
}

function setStatus(statusElement, state, label) {
  statusElement.className = `status-pill ${state}`;
  statusElement.textContent = label;
}

function formatJson(value) {
  return JSON.stringify(value, null, 2);
}
