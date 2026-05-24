// Shared profile, ranking, announcement helpers
function renderCountTable(rows, head1, head2) {
  if (!rows.length) {
    return `<p>No data.</p>`;
  }
  return `
    <table class="data-table">
      <thead>
        <tr>
          <th>${escapeHTML(head1)}</th>
          <th>${escapeHTML(head2)}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => `
          <tr>
            <td>${escapeHTML(row.name)}</td>
            <td>${row.count}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;
}

function renderRecentActivityTable(rows) {
  if (!rows.length) {
    return `<p>No recent activity.</p>`;
  }
  return `
    <table class="data-table">
      <thead>
        <tr>
          <th>Date</th>
          <th>Submissions</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => `
          <tr>
            <td class="mono">${escapeHTML(row.date)}</td>
            <td>${row.count}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;
}

function renderProblemStats(stats) {
  if (!stats) {
    return "";
  }
  return `
    <div class="detail-block">
      <h3>Problem Stats</h3>
      <div class="verdict-summary-grid">
        <div class="verdict-summary-card"><span class="status-pill status-neutral">Submissions</span><strong>${stats.submissions_total}</strong></div>
        <div class="verdict-summary-card"><span class="status-pill status-accepted">Accepted</span><strong>${stats.accepted_submissions}</strong></div>
        <div class="verdict-summary-card"><span class="status-pill status-neutral">Accepted Users</span><strong>${stats.accepted_users}</strong></div>
        <div class="verdict-summary-card"><span class="status-pill status-neutral">Accept Rate</span><strong>${(Number(stats.accepted_rate || 0) * 100).toFixed(1)}%</strong></div>
      </div>
      ${renderCountTable(stats.language_breakdown || [], "Language", "Submissions")}
    </div>
  `;
}



async function renderRankings() {
  app.innerHTML = `<div class="detail-card"><p>Loading ranklist...</p></div>`;
  try {
    const data = await apiFetch('/ranklist?page=1&page_size=50', { method: 'GET' });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Rankings</h1>
          <p class="view-subtitle">Sorted by solved problems, then accepted submissions, then efficiency.</p>
        </div>
      </div>
      <section class="detail-card">
        <table class="data-table">
          <thead>
            <tr>
              <th>Rank</th>
              <th>User</th>
              <th>Student ID</th>
              <th>Solved</th>
              <th>Accepted</th>
              <th>Total</th>
              <th>Last AC</th>
            </tr>
          </thead>
          <tbody>
            ${(data.list || []).map((item) => `
              <tr>
                <td>${item.rank}</td>
                <td>${escapeHTML(item.username)}</td>
                <td class="mono">${escapeHTML(item.userid)}</td>
                <td>${item.solved_count}</td>
                <td>${item.accepted_submissions}</td>
                <td>${item.total_submissions}</td>
                <td class="mono">${escapeHTML(item.last_accepted_at || '-')}</td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load ranklist failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAnnouncements() {
  app.innerHTML = `<div class="detail-card"><p>Loading announcements...</p></div>`;
  try {
    const data = await apiFetch('/announcements?page=1&page_size=20', { method: 'GET' });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Announcements</h1>
          <p class="view-subtitle">Pinned items first, then newest updates.</p>
        </div>
      </div>
      <section class="detail-grid">
        ${(data.list || []).map((item) => `
          <article class="detail-card">
            <div class="view-header">
              <div>
                <h3 style="margin:0;">${escapeHTML(item.title)}</h3>
                <p class="view-subtitle mono">${escapeHTML(item.created_at)}</p>
              </div>
              ${item.is_pinned ? '<span class="status-pill status-pending">Pinned</span>' : ''}
            </div>
            <pre>${escapeHTML(item.content || '')}</pre>
            <div style="margin-top:12px;"><a class="table-link" href="#/announcements/${item.id}">Open</a></div>
          </article>
        `).join('')}
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load announcements failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAnnouncementDetail(id) {
  app.innerHTML = `<div class="detail-card"><p>Loading announcement...</p></div>`;
  try {
    const item = await apiFetch(`/announcements/${id}`, { method: 'GET' });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(item.title)}</h1>
          <p class="view-subtitle mono">${escapeHTML(item.created_at)}</p>
        </div>
        <div>
          ${item.is_pinned ? '<span class="status-pill status-pending">Pinned</span>' : ''}
        </div>
      </div>
      <section class="detail-card">
        <pre>${escapeHTML(item.content || '')}</pre>
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load announcement failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function renderApiDocs() {
  const apiBase = escapeHTML(state.apiBase || `${location.origin}/api`);
  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">SEU OJ API</h1>
        <p class="view-subtitle">Stable interfaces for public data queries and administrator problem data exchange.</p>
      </div>
      <a class="ghost-button" href="#/problems">Browse Problems</a>
    </div>

    <section class="detail-grid">
      <article class="detail-card">
        <h3>Base URL</h3>
        <p class="view-subtitle">Use the same backend API origin configured by the frontend.</p>
        <pre class="mono">${apiBase}</pre>
      </article>
      <article class="detail-card">
        <h3>Response Format</h3>
        <p class="view-subtitle">JSON endpoints use the existing SEU OJ response envelope.</p>
        <pre class="mono">{
  "code": 0,
  "message": "ok",
  "data": {}
}</pre>
      </article>
    </section>

    <section class="detail-card" style="margin-top:18px;">
      <div class="view-header">
        <div>
          <h3 style="margin:0;">Public Read APIs</h3>
          <p class="view-subtitle">These endpoints do not require authentication and never return hidden testcases or source code.</p>
        </div>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Method</th>
            <th>Endpoint</th>
            <th>Purpose</th>
            <th>Query</th>
          </tr>
        </thead>
        <tbody>
          ${renderApiRow("GET", "/public/problems", "Public problem list with accepted and submission counts.", "page, page_size, keyword, difficulty")}
          ${renderApiRow("GET", "/public/problems/:id", "Public problem detail with samples only.", "-")}
          ${renderApiRow("GET", "/public/contests", "Public contest list.", "page, page_size, keyword, status")}
          ${renderApiRow("GET", "/public/contests/:id/ranklist", "Contest ranklist respecting freeze rules.", "-")}
          ${renderApiRow("GET", "/public/submissions", "Public submission summaries without code.", "page, page_size, user_id, problem_id, contest_id, status, language")}
        </tbody>
      </table>
    </section>

    <section class="detail-card" style="margin-top:18px;">
      <div class="view-header">
        <div>
          <h3 style="margin:0;">Admin Problem Data APIs</h3>
          <p class="view-subtitle">These endpoints require administrator authentication and are intended for batch maintenance.</p>
        </div>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Method</th>
            <th>Endpoint</th>
            <th>Purpose</th>
            <th>Body</th>
          </tr>
        </thead>
        <tbody>
          ${renderApiRow("POST", "/admin/problems/import", "Import a complete problem package.", "multipart/form-data: file")}
          ${renderApiRow("GET", "/admin/problems/:id/export", "Export a complete problem package zip.", "-")}
          ${renderApiRow("POST", "/admin/problems/:id/testcases/import", "Import testcase zip into an existing problem.", "multipart/form-data: file, replace, case_type")}
          ${renderApiRow("GET", "/admin/problems/:id/testcases/export", "Export testcase zip for an existing problem.", "-")}
        </tbody>
      </table>
    </section>

    <section class="detail-grid" style="margin-top:18px;">
      <article class="detail-card">
        <h3>Testcase Zip</h3>
        <p class="view-subtitle">Each input file must have a matching output file.</p>
        <pre class="mono">1.in
1.out
2.in
2.out</pre>
      </article>
      <article class="detail-card">
        <h3>Problem Package Zip</h3>
        <p class="view-subtitle">A full package contains problem metadata and testcase files.</p>
        <pre class="mono">problem.json
tests/
  1.in
  1.out
  2.in
  2.out</pre>
      </article>
    </section>

    <section class="detail-card" style="margin-top:18px;">
      <h3>problem.json Example</h3>
      <pre class="mono">{
  "display_id": "1001",
  "title": "A + B Problem",
  "description": "Given two integers, output their sum.",
  "input_desc": "Two integers a and b.",
  "output_desc": "The sum of a and b.",
  "sample_input": "1 2\\n",
  "sample_output": "3\\n",
  "judge_mode": "standard",
  "difficulty": 1,
  "time_limit_ms": 1000,
  "memory_limit_mb": 256,
  "visible": true
}</pre>
    </section>

  `;
}

function renderApiRow(method, endpoint, purpose, query) {
  return `
    <tr>
      <td><span class="status-pill status-neutral">${escapeHTML(method)}</span></td>
      <td class="mono">${escapeHTML(endpoint)}</td>
      <td>${escapeHTML(purpose)}</td>
      <td class="mono">${escapeHTML(query)}</td>
    </tr>
  `;
}
