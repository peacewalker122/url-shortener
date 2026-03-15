import { check } from 'k6';
import http from 'k6/http';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const PROFILE = __ENV.PROFILE || 'balanced';

const DEFAULT_HEADERS = { 'Content-Type': 'application/json' };

function createScenarioConfig() {
  return {
    executor: 'constant-arrival-rate',
    duration: __ENV.DURATION || '30s',
    rate: Number(__ENV.RATE || 20),
    timeUnit: '1s',
    preAllocatedVUs: Number(__ENV.PRE_ALLOCATED_VUS || 20),
    maxVUs: Number(__ENV.MAX_VUS || 100),
  };
}

function createThresholds() {
  return {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    'http_req_failed{endpoint:post_shorten}': ['rate<0.01'],
    'http_req_failed{endpoint:get_shorten_key}': ['rate<0.01'],
    'http_req_duration{endpoint:post_shorten}': ['p(95)<500', 'p(99)<1000'],
    'http_req_duration{endpoint:get_shorten_key}': ['p(95)<500', 'p(99)<1000'],
  };
}

export const options = {
  scenarios: {
    url_shortener_flow: createScenarioConfig(),
  },
  thresholds: createThresholds(),
};

function createShortUrl(longUrl) {
  const payload = JSON.stringify({ url: longUrl });
  const response = http.post(`${BASE_URL}/shorten`, payload, {
    headers: DEFAULT_HEADERS,
    tags: { endpoint: 'post_shorten' },
  });

  const postOk = check(response, {
    'POST /shorten status is 200': (r) => r.status === 200,
    'POST /shorten returns short url key': (r) => {
      const body = r.json();
      return !!body && typeof body.url === 'string' && body.url.length > 0;
    },
  });

  if (!postOk) {
    return null;
  }

  return response.json().url;
}

function resolveShortUrl(shortKey, expectedLongUrl) {
  const response = http.get(`${BASE_URL}/${shortKey}`, {
    redirects: 0,
    tags: { endpoint: 'get_shorten_key' },
  });

  check(response, {
    'GET /{key} status is 302': (r) => r.status === 302,
    'GET /{key} redirects to original url': (r) => r.headers.Location === expectedLongUrl,
  });
}

export function setup() {
  if (PROFILE !== 'read-heavy') {
    return { profile: PROFILE, seeded: [] };
  }

  const seedCount = Number(__ENV.SEED_COUNT || 100);
  const seeded = [];

  for (let i = 0; i < seedCount; i += 1) {
    const longUrl = `https://example.com/seed/${i}`;
    const shortKey = createShortUrl(longUrl);
    if (shortKey) {
      seeded.push({ shortKey, longUrl });
    }
  }

  return { profile: PROFILE, seeded };
}

export default function (data) {
  if (data.profile === 'read-heavy') {
    if (data.seeded.length === 0) {
      return;
    }

    // 10:1 read/write mix per iteration (10 GET + 1 POST).
    for (let i = 0; i < 10; i += 1) {
      const index = Math.floor(Math.random() * data.seeded.length);
      const selected = data.seeded[index];
      resolveShortUrl(selected.shortKey, selected.longUrl);
    }

    const longUrl = `https://example.com/read-heavy/${__VU}-${__ITER}`;
    const shortKey = createShortUrl(longUrl);
    if (shortKey) {
      data.seeded.push({ shortKey, longUrl });
    }
    return;
  }

  // Balanced flow: 1 write then 1 read.
  const longUrl = `https://example.com/resource/${__VU}-${__ITER}`;
  const shortKey = createShortUrl(longUrl);
  if (!shortKey) {
    return;
  }
  resolveShortUrl(shortKey, longUrl);
}
