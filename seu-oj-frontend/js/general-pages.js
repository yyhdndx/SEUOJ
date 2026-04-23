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


