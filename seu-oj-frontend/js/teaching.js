// Teaching domain pages and helpers
function isTeacherUser() {
  return !!state.user && (state.user.role === "teacher" || state.user.role === "admin");
}

function toRFC3339(localValue) {
  if (!localValue) return null;
  const date = new Date(localValue);
  return Number.isNaN(date.getTime()) ? null : date.toISOString();
}
function toDatetimeLocalValue(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
function parseProblemIDs(text) {
  return String(text || "")
    .split(/[\s,]+/)
    .map((item) => Number(item.trim()))
    .filter((value) => Number.isFinite(value) && value > 0);
}
function teachingVisibilityClass(visibility) {
  if (visibility === "public") return "status-accepted";
  if (visibility === "class") return "status-pending";
  return "status-neutral";
}

function teachingStatusClass(status) {
  if (status === "active" || status === "open" || status === "running") return "status-accepted";
  if (status === "upcoming" || status === "pending") return "status-pending";
  if (status === "archived" || status === "closed" || status === "ended") return "status-neutral";
  return "status-neutral";
}

function renderTeachingSummaryCards(items) {
  if (!items || !items.length) {
    return `
      <div class="teaching-summary-grid">
        <article class="teaching-summary-card"><span class="teaching-summary-label">Items</span><strong>0</strong></article>
      </div>
    `;
  }

  const totalProblems = items.reduce((sum, item) => sum + Number(item.problem_count || 0), 0);
  const totalAssignments = items.reduce((sum, item) => sum + Number(item.assignment_count || 0), 0);
  const activeItems = items.filter((item) => String(item.status || "").toLowerCase() === "active").length;
  return `
    <div class="teaching-summary-grid">
      <article class="teaching-summary-card"><span class="teaching-summary-label">Items</span><strong>${items.length}</strong></article>
      <article class="teaching-summary-card"><span class="teaching-summary-label">Active</span><strong>${activeItems}</strong></article>
      <article class="teaching-summary-card"><span class="teaching-summary-label">Problems</span><strong>${totalProblems}</strong></article>
      <article class="teaching-summary-card"><span class="teaching-summary-label">Assignments</span><strong>${totalAssignments}</strong></article>
    </div>
  `;
}

function renderTeachingMetaList(items) {
  return `
    <div class="teaching-meta-grid">
      ${items.map((item) => `
        <div class="teaching-meta-card">
          <span class="teaching-meta-label">${escapeHTML(item.label)}</span>
          <span class="teaching-meta-value ${item.mono ? "mono" : ""}">${escapeHTML(item.value)}</span>
        </div>
      `).join("")}
    </div>
  `;
}

function renderTeachingEmpty(message) {
  return `<div class="teaching-empty">${escapeHTML(message)}</div>`;
}

function renderAssignmentProgressPill(item) {
  const solved = Number(item.solved_count || 0);
  const total = Number(item.total_count || 0);
  if (total > 0 && solved >= total) {
    return `<span class="status-pill status-accepted">${solved}/${total}</span>`;
  }
  if (solved > 0) {
    return `<span class="status-pill status-pending">${solved}/${total}</span>`;
  }
  return `<span class="status-pill status-neutral">${solved}/${total}</span>`;
}


async function renderPlaylists() {
  app.innerHTML = `<div class="detail-card"><p>Loading playlists...</p></div>`;
  try {
    const data = await apiFetch("/playlists?page=1&page_size=50", { method: "GET" });
    const list = data.list || [];
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Playlists</h1>
          <p class="view-subtitle">Reusable problem sets for practice, homework, and exams.</p>
        </div>
      </div>
      ${renderTeachingSummaryCards(list.map((item) => ({ ...item, assignment_count: 0, status: item.visibility === 'public' ? 'active' : 'archived' })))}
      <section class="teaching-card-grid" style="margin-top:18px;">
        ${list.map((item) => `
          <article class="detail-card teaching-list-card">
            <div class="view-header compact">
              <div>
                <h3 class="teaching-card-title"><a class="table-link" href="#/playlists/${item.id}">${escapeHTML(item.title)}</a></h3>
                <p class="view-subtitle">Playlist #${item.id}</p>
              </div>
              <span class="status-pill ${teachingVisibilityClass(item.visibility)}">${escapeHTML(item.visibility)}</span>
            </div>
            <p class="teaching-card-copy">${escapeHTML((item.description || 'No description.').slice(0, 120))}${item.description && item.description.length > 120 ? '...' : ''}</p>
            ${renderTeachingMetaList([
              { label: 'Problems', value: String(item.problem_count) },
              { label: 'Created', value: item.created_at, mono: true },
            ])}
            <div style="margin-top:14px;"><a class="ghost-button" href="#/playlists/${item.id}">Open Playlist</a></div>
          </article>
        `).join("")}
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load playlists failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderPlaylistDetail(id) {
  app.innerHTML = `<div class="detail-card"><p>Loading playlist...</p></div>`;
  try {
    const detail = await apiFetch(`/playlists/${id}`, { method: "GET" });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">Reusable set for training, homework, and exam assembly.</p>
        </div>
        <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
          <span class="status-pill ${teachingVisibilityClass(detail.visibility)}">${escapeHTML(detail.visibility)}</span>
          <a class="ghost-button" href="#/playlists">Back</a>
        </div>
      </div>
      ${renderTeachingMetaList([
        { label: 'Problems', value: String((detail.problems || []).length) },
        { label: 'Created By', value: String(detail.created_by) },
        { label: 'Created', value: detail.created_at, mono: true },
        { label: 'Updated', value: detail.updated_at, mono: true },
      ])}
      <section class="detail-grid" style="margin-top:18px;">
        <article class="detail-card">
          <h3>Description</h3>
          ${renderMarkdownBlock(detail.description || "No description.")}
        </article>
        <aside class="detail-card">
          <h3>Problems</h3>
          ${(detail.problems || []).length ? `
            <table class="data-table">
              <thead><tr><th>#</th><th>Display</th><th>Title</th></tr></thead>
              <tbody>
                ${(detail.problems || []).map((item) => `
                  <tr>
                    <td>${item.display_order}</td>
                    <td class="mono">${escapeHTML(item.display_id || "-")}</td>
                    <td><a class="table-link" href="#/problems/${item.problem_id}">${escapeHTML(item.title)}</a></td>
                  </tr>
                `).join("")}
              </tbody>
            </table>
          ` : renderTeachingEmpty('No problems in this playlist.')}
        </aside>
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load playlist failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderClasses() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login to access classes.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading classes...</p></div>`;
  try {
    const data = await apiFetch("/classes/my", { method: "GET" });
    const list = data.list || [];
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">My Classes</h1>
          <p class="view-subtitle">Join a class to receive homework playlists and exam assignments.</p>
        </div>
      </div>
      ${renderTeachingSummaryCards(list)}
      <section class="detail-grid" style="margin-top:18px;">
        <article class="detail-card">
          <h3>Join Class</h3>
          <form id="join-class-form">
            <label class="field-label">Join Code</label>
            <input class="text-input" name="join_code" placeholder="e.g. ABCD1234" required />
            <div class="view-subtitle" style="margin-top:8px;">Teachers distribute join codes to build classroom membership.</div>
            <div style="margin-top:14px;"><button class="primary-button" type="submit">Join</button></div>
          </form>
        </article>
        <aside class="detail-card">
          <h3>Current Classes</h3>
          ${list.length ? `
            <div class="teaching-stack">
              ${list.map((item) => `
                <article class="teaching-list-card">
                  <div class="view-header compact">
                    <div>
                      <h3 class="teaching-card-title"><a class="table-link" href="#/classes/${item.id}">${escapeHTML(item.name)}</a></h3>
                      <p class="view-subtitle">Teacher ${escapeHTML(item.teacher_name)} / role ${escapeHTML(item.member_role || 'student')}</p>
                    </div>
                    <span class="status-pill ${teachingStatusClass(item.status)}">${escapeHTML(item.status)}</span>
                  </div>
                  ${renderTeachingMetaList([
                    { label: 'Assignments', value: String(item.assignment_count) },
                    { label: 'Members', value: String(item.member_count) },
                    { label: 'Created', value: item.created_at, mono: true },
                  ])}
                </article>
              `).join("")}
            </div>
          ` : renderTeachingEmpty('No classes yet.')}
        </aside>
      </section>
    `;
    document.getElementById("join-class-form")?.addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        const detail = await apiFetch("/classes/join", { method: "POST", body: JSON.stringify({ join_code: form.get("join_code") }) });
        setFlash("Class joined", false);
        location.hash = `#/classes/${detail.id}`;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load classes failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderClassDetail(id) {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login to access class detail.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading class...</p></div>`;
  try {
    const detail = await apiFetch(`/classes/${id}`, { method: "GET" });
    const assignments = detail.assignments || [];
    const openCount = assignments.filter((item) => item.status === 'open').length;
    const closedCount = assignments.filter((item) => item.status === 'closed').length;
    const completedCount = assignments.filter((item) => Number(item.total_count || 0) > 0 && Number(item.solved_count || 0) >= Number(item.total_count || 0)).length;
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.name)}</h1>
          <p class="view-subtitle">Teacher ${escapeHTML(detail.teacher_name)} / classroom workspace</p>
        </div>
        <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
          <span class="status-pill ${teachingStatusClass(detail.status)}">${escapeHTML(detail.status)}</span>
          <a class="ghost-button" href="#/classes">Back</a>
        </div>
      </div>
      <section class="teaching-analytics-grid">
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Assignments</span>
          <strong>${assignments.length}</strong>
          <span class="view-subtitle">${openCount} open / ${closedCount} closed</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Completed</span>
          <strong>${completedCount}</strong>
          <span class="view-subtitle">Assignments fully solved</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Teacher</span>
          <strong>${escapeHTML(detail.teacher_name)}</strong>
          <span class="view-subtitle">Class owner</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Members</span>
          <strong>${(detail.members || []).length || '-'}</strong>
          <span class="view-subtitle">Visible in your classroom view</span>
        </article>
      </section>
      <section class="detail-grid" style="margin-top:18px; align-items:start;">
        <article class="detail-card">
          <h3>Description</h3>
          ${renderMarkdownBlock(detail.description || "No description.")}
        </article>
        <aside class="detail-card">
          <h3>Assignment Snapshot</h3>
          ${assignments.length ? `
            <div class="teaching-stack">
              ${assignments.slice(0, 5).map((item) => `
                <article class="teaching-ranking-card">
                  <div class="view-header compact">
                    <div>
                      <h4 class="teaching-card-title"><a class="table-link" href="#/assignments/${item.id}">${escapeHTML(item.title)}</a></h4>
                      <p class="view-subtitle">${escapeHTML(item.type)} / ${escapeHTML(item.status)}</p>
                    </div>
                    ${renderAssignmentProgressPill(item)}
                  </div>
                  ${renderTeachingMetaList([
                    { label: 'Window', value: `${item.start_at || '-'} -> ${item.due_at || '-'}`, mono: true },
                    { label: 'Solved', value: `${item.solved_count}/${item.total_count}` },
                  ])}
                </article>
              `).join("")}
            </div>
          ` : renderTeachingEmpty('No assignments yet.')}
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px; overflow:auto;">
        <div class="view-header compact">
          <div>
            <h3>Assignments</h3>
            <p class="view-subtitle">Track homework and exams assigned to this class.</p>
          </div>
        </div>
        ${assignments.length ? `
          <table class="data-table assignment-analytics-table">
            <thead><tr><th>Title</th><th>Type</th><th>Status</th><th>Progress</th><th>Window</th><th>Action</th></tr></thead>
            <tbody>
              ${assignments.map((item) => `
                <tr>
                  <td><strong>${escapeHTML(item.title)}</strong></td>
                  <td><span class="status-pill ${teachingStatusClass(item.type)}">${escapeHTML(item.type)}</span></td>
                  <td><span class="status-pill ${teachingStatusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                  <td>${renderAssignmentProgressPill(item)}</td>
                  <td class="mono">${escapeHTML(`${item.start_at || '-'} -> ${item.due_at || '-'}`)}</td>
                  <td><a class="table-link" href="#/assignments/${item.id}">Open</a></td>
                </tr>
              `).join("")}
            </tbody>
          </table>
        ` : renderTeachingEmpty('No assignments yet.')}
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load class failed: ${escapeHTML(err.message)}</p></div>`;
  }
}


﻿async function renderAssignmentDetail(id) {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login to access assignment detail.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading assignment...</p></div>`;
  try {
    const detail = await apiFetch(`/assignments/${id}`, { method: "GET" });
    const problems = detail.problems || [];
    const solvedProblems = problems.filter((item) => item.solved).length;
    const unsolvedProblems = problems.filter((item) => !item.solved);
    const nextProblem = unsolvedProblems[0] || problems[0] || null;
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">${escapeHTML(detail.class_name)} / ${escapeHTML(detail.playlist_title)}</p>
        </div>
        <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
          <span class="status-pill ${teachingStatusClass(detail.type)}">${escapeHTML(detail.type)}</span>
          <span class="status-pill ${teachingStatusClass(detail.status)}">${escapeHTML(detail.status)}</span>
          <a class="ghost-button" href="#/classes/${detail.class_id}">Back to Class</a>
          ${nextProblem ? `<a class="ghost-button" href="#/problems/${nextProblem.problem_id}">Continue Work</a>` : ''}
        </div>
      </div>
      <section class="teaching-analytics-grid">
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Solved</span>
          <strong>${detail.solved_count}/${detail.total_count}</strong>
          <span class="view-subtitle">Assignment progress</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Next Focus</span>
          <strong>${escapeHTML(nextProblem ? (nextProblem.display_id || String(nextProblem.display_order)) : '-')}</strong>
          <span class="view-subtitle">${escapeHTML(nextProblem ? nextProblem.title : 'No problem available')}</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Window</span>
          <strong>${detail.status === 'open' ? 'Open' : detail.status === 'closed' ? 'Closed' : 'Upcoming'}</strong>
          <span class="view-subtitle mono">${escapeHTML(`${detail.start_at || '-'} -> ${detail.due_at || '-'}`)}</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Unsolved</span>
          <strong>${detail.total_count - solvedProblems}</strong>
          <span class="view-subtitle">Remaining problems</span>
        </article>
      </section>
      <section class="detail-grid" style="margin-top:18px; align-items:start;">
        <article class="detail-card">
          <h3>Description</h3>
          ${renderMarkdownBlock(detail.description || "No description.")}
        </article>
        <aside class="detail-card">
          <h3>Quick Progress</h3>
          ${problems.length ? `
            <div class="teaching-stack">
              ${problems.map((item) => `
                <article class="teaching-ranking-card">
                  <div class="view-header compact">
                    <div>
                      <h4 class="teaching-card-title">${escapeHTML(item.display_id || String(item.display_order))} · ${escapeHTML(item.title)}</h4>
                      <p class="view-subtitle">Order ${item.display_order}</p>
                    </div>
                    ${item.solved ? `<span class="status-pill status-accepted">Accepted</span>` : renderProblemStatusPill(item.last_status || null)}
                  </div>
                  <div style="margin-top:10px;"><a class="ghost-button" href="#/problems/${item.problem_id}">Open Problem</a></div>
                </article>
              `).join("")}
            </div>
          ` : renderTeachingEmpty('No problems in this assignment.')}
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px; overflow:auto;">
        <div class="view-header compact">
          <div>
            <h3>Problem Checklist</h3>
            <p class="view-subtitle">Open each problem and keep progressing until the entire assignment is accepted.</p>
          </div>
        </div>
        <table class="data-table assignment-analytics-table">
          <thead><tr><th>#</th><th>Problem</th><th>Status</th><th>Action</th></tr></thead>
          <tbody>
            ${problems.map((item) => `
              <tr>
                <td>${item.display_order}</td>
                <td><strong>${escapeHTML(item.title)}</strong></td>
                <td>${item.solved ? `<span class="status-pill status-accepted">Accepted</span>` : renderProblemStatusPill(item.last_status || null)}</td>
                <td><a class="table-link" href="#/problems/${item.problem_id}">Open</a></td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load assignment failed: ${escapeHTML(err.message)}</p></div>`;
  }
}


async function renderTeacherPlaylists() {
  if (!isTeacherUser()) {
    app.innerHTML = `<div class="detail-card"><p>Teacher permission required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading teacher playlists...</p></div>`;
  try {
    const params = getHashQueryParams();
    const keyword = String(params.keyword || '').trim();
    const querySuffix = keyword ? `&keyword=${encodeURIComponent(keyword)}` : '';
    const [playlists, problems] = await Promise.all([
      apiFetch(`/teacher/playlists?page=1&page_size=100${querySuffix}`, { method: 'GET' }),
      apiFetch('/problems?page=1&page_size=200', { method: 'GET' }),
    ]);
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Teaching Playlists</h1>
          <p class="view-subtitle">Create reusable problem sets for class homework and exams.</p>
        </div>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          <h3>Create Playlist</h3>
          <form id="teacher-playlist-form">
            <label class="field-label">Title</label>
            <input class="text-input" name="title" required />
            <label class="field-label">Visibility</label>
            <select class="select-input" name="visibility">
              <option value="public">public</option>
              <option value="private">private</option>
              <option value="class">class</option>
            </select>
            <label class="field-label">Description</label>
            <textarea class="text-area" name="description" rows="6"></textarea>
            <label class="field-label">Problem IDs (ordered, split by comma or space)</label>
            <input class="text-input" name="problem_ids" placeholder="3, 4, 5" required />
            <div class="view-subtitle" style="margin-top:8px;">Available problems: ${(problems.list || []).map((item) => `#${item.id} ${escapeHTML(item.title)}`).join(' / ')}</div>
            <div style="margin-top:14px;"><button class="primary-button" type="submit">Create Playlist</button></div>
          </form>
        </article>
        <aside class="detail-card">
          <div class="view-header compact">
            <div>
              <h3>Existing Playlists</h3>
              <p class="view-subtitle">Search and maintain your reusable sets.</p>
            </div>
            <span class="status-pill status-neutral">${(playlists.list || []).length} items</span>
          </div>
          <form id="teacher-playlist-search-form" style="margin-bottom:12px; display:flex; gap:10px; align-items:center;">
            <input class="text-input" name="keyword" placeholder="Search title" value="${escapeHTML(keyword)}" />
            <button class="ghost-button" type="submit">Search</button>
            ${keyword ? '<a class="ghost-button" href="#/teacher/playlists">Clear</a>' : ''}
          </form>
          <table class="data-table">
            <thead><tr><th>Title</th><th>Visibility</th><th>Problems</th><th>Actions</th></tr></thead>
            <tbody>
              ${(playlists.list || []).map((item) => `
                <tr>
                  <td>${escapeHTML(item.title)}</td>
                  <td>${escapeHTML(item.visibility)}</td>
                  <td>${item.problem_count}</td>
                  <td><a class="table-link" href="#/teacher/playlists/${item.id}">Open</a></td>
                </tr>
              `).join('')}
            </tbody>
          </table>
          ${(playlists.list || []).length ? '' : renderTeachingEmpty('No playlist matched the current keyword.')}
        </aside>
      </section>
    `;
    document.getElementById('teacher-playlist-form')?.addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const ids = parseProblemIDs(form.get('problem_ids'));
      try {
        await apiFetch('/teacher/playlists', {
          method: 'POST',
          body: JSON.stringify({
            title: form.get('title'),
            visibility: form.get('visibility'),
            description: form.get('description'),
            problems: ids.map((problemID, index) => ({ problem_id: problemID, display_order: index + 1 })),
          }),
        });
        setFlash('Playlist created', false);
        return renderTeacherPlaylists();
      } catch (err) {
        setFlash(err.message, true);
      }
    });
    document.getElementById('teacher-playlist-search-form')?.addEventListener('submit', (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const nextKeyword = String(form.get('keyword') || '').trim();
      location.hash = nextKeyword ? `#/teacher/playlists?keyword=${encodeURIComponent(nextKeyword)}` : '#/teacher/playlists';
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load teacher playlists failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderTeacherPlaylistDetail(id) {
  if (!isTeacherUser()) {
    app.innerHTML = `<div class="detail-card"><p>Teacher permission required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading teacher playlist...</p></div>`;
  try {
    const [detail, problems] = await Promise.all([
      apiFetch(`/teacher/playlists/${id}`, { method: 'GET' }),
      apiFetch('/problems?page=1&page_size=200', { method: 'GET' }),
    ]);
    const problemIDs = (detail.problems || []).map((item) => item.problem_id).join(', ');
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">${escapeHTML(detail.visibility)} playlist / ${detail.problems?.length || 0} problems</p>
        </div>
        <a class="ghost-button" href="#/teacher/playlists">Back</a>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          <h3>Edit Playlist</h3>
          <form id="teacher-playlist-edit-form">
            <label class="field-label">Title</label>
            <input class="text-input" name="title" value="${escapeHTML(detail.title)}" required />
            <label class="field-label">Visibility</label>
            <select class="select-input" name="visibility">
              <option value="public" ${detail.visibility === 'public' ? 'selected' : ''}>public</option>
              <option value="private" ${detail.visibility === 'private' ? 'selected' : ''}>private</option>
              <option value="class" ${detail.visibility === 'class' ? 'selected' : ''}>class</option>
            </select>
            <label class="field-label">Description</label>
            <textarea class="text-area" name="description" rows="8">${escapeHTML(detail.description || '')}</textarea>
            <label class="field-label">Problem IDs (ordered)</label>
            <input class="text-input" name="problem_ids" value="${escapeHTML(problemIDs)}" required />
            <div class="view-subtitle" style="margin-top:8px;">Available problems: ${(problems.list || []).map((item) => `#${item.id} ${escapeHTML(item.title)}`).join(' / ')}</div>
            <div style="margin-top:14px; display:flex; gap:10px; flex-wrap:wrap;">
              <button class="primary-button" type="submit">Save Playlist</button>
              <a class="ghost-button" href="#/playlists/${detail.id}">Open Public View</a>
              <button class="ghost-button" id="teacher-playlist-delete-button" type="button">Delete Playlist</button>
            </div>
          </form>
        </article>
        <aside class="detail-card">
          <h3>Problems</h3>
          <table class="data-table">
            <thead><tr><th>#</th><th>Display</th><th>Title</th></tr></thead>
            <tbody>
              ${(detail.problems || []).map((item) => `
                <tr>
                  <td>${item.display_order}</td>
                  <td class="mono">${escapeHTML(item.display_id || '-')}</td>
                  <td><a class="table-link" href="#/problems/${item.problem_id}">${escapeHTML(item.title)}</a></td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </aside>
      </section>
    `;
    document.getElementById('teacher-playlist-edit-form')?.addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const ids = parseProblemIDs(form.get('problem_ids'));
      try {
        await apiFetch(`/teacher/playlists/${id}`, {
          method: 'PUT',
          body: JSON.stringify({
            title: form.get('title'),
            visibility: form.get('visibility'),
            description: form.get('description'),
            problems: ids.map((problemID, index) => ({ problem_id: problemID, display_order: index + 1 })),
          }),
        });
        setFlash('Playlist updated', false);
        return renderTeacherPlaylistDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });
    document.getElementById('teacher-playlist-delete-button')?.addEventListener('click', async () => {
      if (!window.confirm(`Delete playlist #${id}?`)) return;
      try {
        await apiFetch(`/teacher/playlists/${id}`, { method: 'DELETE' });
        setFlash('Playlist deleted', false);
        location.hash = '#/teacher/playlists';
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load teacher playlist failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderTeacherClasses() {
  if (!isTeacherUser()) {
    app.innerHTML = `<div class="detail-card"><p>Teacher permission required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading teacher classes...</p></div>`;
  try {
    const params = getHashQueryParams();
    const keyword = String(params.keyword || '').trim().toLowerCase();
    const status = String(params.status || 'all').trim().toLowerCase();
    const data = await apiFetch('/teacher/classes', { method: 'GET' });
    const allClasses = data.list || [];
    const filteredClasses = allClasses.filter((item) => {
      const matchesKeyword = !keyword || String(item.name || '').toLowerCase().includes(keyword) || String(item.description || '').toLowerCase().includes(keyword);
      const matchesStatus = status === 'all' || String(item.status || '').toLowerCase() === status;
      return matchesKeyword && matchesStatus;
    });
    const activeCount = allClasses.filter((item) => String(item.status || '').toLowerCase() === 'active').length;
    const archivedCount = allClasses.filter((item) => String(item.status || '').toLowerCase() === 'archived').length;
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Teaching Classes</h1>
          <p class="view-subtitle">Create classes and publish assignment playlists.</p>
        </div>
      </div>
      <section class="teaching-summary-grid" style="margin-bottom:18px;">
        <article class="teaching-summary-card"><span class="teaching-summary-label">Classes</span><strong>${allClasses.length}</strong></article>
        <article class="teaching-summary-card"><span class="teaching-summary-label">Active</span><strong>${activeCount}</strong></article>
        <article class="teaching-summary-card"><span class="teaching-summary-label">Archived</span><strong>${archivedCount}</strong></article>
        <article class="teaching-summary-card"><span class="teaching-summary-label">Showing</span><strong>${filteredClasses.length}</strong></article>
      </section>
      <section class="detail-grid">
        <article class="detail-card">
          <h3>Create Class</h3>
          <form id="teacher-class-form">
            <label class="field-label">Name</label>
            <input class="text-input" name="name" required />
            <label class="field-label">Description</label>
            <textarea class="text-area" name="description" rows="6"></textarea>
            <div style="margin-top:14px;"><button class="primary-button" type="submit">Create Class</button></div>
          </form>
        </article>
        <aside class="detail-card">
          <div class="view-header compact">
            <div>
              <h3>Existing Classes</h3>
              <p class="view-subtitle">Search and filter your class space.</p>
            </div>
            <span class="status-pill status-neutral">${filteredClasses.length} shown</span>
          </div>
          <form id="teacher-class-filter-form" style="margin-bottom:12px; display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
            <input class="text-input" name="keyword" placeholder="Search class name" value="${escapeHTML(keyword)}" />
            <select class="select-input" name="status">
              <option value="all" ${status === 'all' ? 'selected' : ''}>all status</option>
              <option value="active" ${status === 'active' ? 'selected' : ''}>active</option>
              <option value="archived" ${status === 'archived' ? 'selected' : ''}>archived</option>
            </select>
            <button class="ghost-button" type="submit">Apply</button>
            ${(keyword || status !== 'all') ? '<a class="ghost-button" href="#/teacher/classes">Clear</a>' : ''}
          </form>
          <table class="data-table">
            <thead><tr><th>Name</th><th>Status</th><th>Join Code</th><th>Members</th><th></th></tr></thead>
            <tbody>
              ${filteredClasses.map((item) => `
                <tr>
                  <td>
                    <strong>${escapeHTML(item.name)}</strong>
                    <div class="view-subtitle">${escapeHTML((item.description || '').slice(0, 60) || 'No description.')}</div>
                  </td>
                  <td><span class="status-pill ${teachingStatusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                  <td class="mono">${escapeHTML(item.join_code || '-')}</td>
                  <td>${item.member_count}</td>
                  <td>
                    <a class="table-link" href="#/teacher/classes/${item.id}">Open</a>
                    <span class="table-separator">/</span>
                    <button class="link-like teacher-copy-join-code" type="button" data-join-code="${escapeHTML(item.join_code || '')}">Copy Code</button>
                  </td>
                </tr>
              `).join('')}
            </tbody>
          </table>
          ${filteredClasses.length ? '' : renderTeachingEmpty('No class matched the current filters.')}
        </aside>
      </section>
    `;
    document.getElementById('teacher-class-form')?.addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        const created = await apiFetch('/teacher/classes', {
          method: 'POST',
          body: JSON.stringify({ name: form.get('name'), description: form.get('description') }),
        });
        setFlash('Class created', false);
        location.hash = `#/teacher/classes/${created.id}`;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
    document.getElementById('teacher-class-filter-form')?.addEventListener('submit', (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const nextKeyword = String(form.get('keyword') || '').trim();
      const nextStatus = String(form.get('status') || 'all').trim();
      const query = new URLSearchParams();
      if (nextKeyword) query.set('keyword', nextKeyword);
      if (nextStatus && nextStatus !== 'all') query.set('status', nextStatus);
      const suffix = query.toString();
      location.hash = suffix ? `#/teacher/classes?${suffix}` : '#/teacher/classes';
    });
    document.querySelectorAll('.teacher-copy-join-code').forEach((button) => {
      button.addEventListener('click', async () => {
        const joinCode = button.dataset.joinCode || '';
        if (!joinCode) return;
        try {
          await navigator.clipboard.writeText(joinCode);
          setFlash(`Join code copied: ${joinCode}`, false);
        } catch (_) {
          setFlash(`Join code: ${joinCode}`, false);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load teacher classes failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderTeacherClassDetail(id) {
  if (!isTeacherUser()) {
    app.innerHTML = `<div class="detail-card"><p>Teacher permission required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading teacher class detail...</p></div>`;
  try {
    const [detail, playlists, analytics] = await Promise.all([
      apiFetch(`/teacher/classes/${id}`, { method: 'GET' }),
      apiFetch('/teacher/playlists?page=1&page_size=100', { method: 'GET' }),
      apiFetch(`/teacher/classes/${id}/analytics`, { method: 'GET' }),
    ]);
    const query = getHashQueryParams();
    const memberKeyword = (query.get('member_keyword') || '').trim().toLowerCase();
    const memberStatusFilter = (query.get('member_status') || 'all').trim().toLowerCase();
    const assignmentStatusFilter = (query.get('assignment_status') || 'all').trim().toLowerCase();
    const filteredMembers = (detail.members || []).filter((item) => {
      const haystack = `${item.username || ''} ${item.userid || ''}`.toLowerCase();
      const matchesKeyword = !memberKeyword || haystack.includes(memberKeyword);
      const matchesStatus = memberStatusFilter === 'all' || (item.status || '').toLowerCase() === memberStatusFilter;
      return matchesKeyword && matchesStatus;
    });
    const filteredAssignments = (analytics.assignments || []).filter((item) => {
      return assignmentStatusFilter === 'all' || (item.status || '').toLowerCase() === assignmentStatusFilter;
    });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.name)}</h1>
          <p class="view-subtitle">Join code: ${escapeHTML(detail.join_code || '-')} / teacher: ${escapeHTML(detail.teacher_name)}</p>
        </div>
        <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
          <span class="status-pill ${teachingStatusClass(detail.status)}">${escapeHTML(detail.status)}</span>
          <a class="ghost-button" href="#/teacher/classes">Back</a>
        </div>
      </div>
      <section class="teaching-analytics-grid">
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Members</span>
          <strong>${analytics.member_count}</strong>
          <span class="view-subtitle">${analytics.active_member_count} active</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Assignments</span>
          <strong>${analytics.assignment_count}</strong>
          <span class="view-subtitle">${analytics.completed_assignments} with completions</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Unique Problems</span>
          <strong>${analytics.unique_problem_count}</strong>
          <span class="view-subtitle">Across all playlists</span>
        </article>
        <article class="teaching-analytics-card">
          <span class="teaching-summary-label">Completion Rate</span>
          <strong>${analytics.completion_rate}%</strong>
          <span class="view-subtitle">Assignment-level completion</span>
        </article>
      </section>
      <section class="detail-grid" style="margin-top:18px; align-items:start;">
        <article class="detail-card">
          <h3>Class Settings</h3>
          <form id="teacher-class-edit-form">
            <label class="field-label">Name</label>
            <input class="text-input" name="name" value="${escapeHTML(detail.name)}" required />
            <label class="field-label">Status</label>
            <select class="select-input" name="status">
              <option value="active" ${detail.status === 'active' ? 'selected' : ''}>active</option>
              <option value="archived" ${detail.status === 'archived' ? 'selected' : ''}>archived</option>
            </select>
            <label class="field-label">Description</label>
            <textarea class="text-area" name="description" rows="6">${escapeHTML(detail.description || '')}</textarea>
            <div style="margin-top:14px;"><button class="primary-button" type="submit">Save Class</button></div>
          </form>
          <h3 style="margin-top:18px;">Create Assignment</h3>
          <form id="teacher-assignment-form">
            <label class="field-label">Title</label>
            <input class="text-input" name="title" required />
            <label class="field-label">Type</label>
            <select class="select-input" name="type">
              <option value="homework">homework</option>
              <option value="exam">exam</option>
            </select>
            <label class="field-label">Playlist</label>
            <select class="select-input" name="playlist_id">
              ${(playlists.list || []).map((item) => `<option value="${item.id}">${escapeHTML(item.title)} (${escapeHTML(item.visibility)})</option>`).join('')}
            </select>
            <label class="field-label">Description</label>
            <textarea class="text-area" name="description" rows="5"></textarea>
            <label class="field-label">Start At</label>
            <input class="text-input" type="datetime-local" name="start_at" />
            <label class="field-label">Due At</label>
            <input class="text-input" type="datetime-local" name="due_at" />
            <div style="margin-top:14px;"><button class="primary-button" type="submit">Create Assignment</button></div>
          </form>
        </article>
        <aside class="detail-card">
          <div class="view-header compact">
            <div>
              <h3>Members</h3>
              <p class="view-subtitle">Manage classroom roles and status.</p>
            </div>
            <span class="status-pill status-neutral">${filteredMembers.length} / ${(detail.members || []).length}</span>
          </div>
          <form id="teacher-member-filter-form" class="teaching-filter-bar" style="margin-top:14px; margin-bottom:12px;">
            <input class="text-input" name="member_keyword" placeholder="Search username or userid" value="${escapeHTML(query.get('member_keyword') || '')}" />
            <select class="select-input" name="member_status">
              <option value="all" ${memberStatusFilter === 'all' ? 'selected' : ''}>all status</option>
              <option value="active" ${memberStatusFilter === 'active' ? 'selected' : ''}>active</option>
              <option value="removed" ${memberStatusFilter === 'removed' ? 'selected' : ''}>removed</option>
            </select>
            <button class="ghost-button" type="submit">Apply</button>
          </form>
          ${filteredMembers.length ? `
            <table class="data-table">
              <thead><tr><th>User</th><th>Role</th><th>Status</th><th>Action</th></tr></thead>
              <tbody>
                ${filteredMembers.map((item) => `
                  <tr>
                    <td>${escapeHTML(item.username)} <span class="mono">${escapeHTML(item.userid)}</span></td>
                    <td>
                      <select class="select-input class-member-role" data-user-id="${item.user_id}">
                        <option value="student" ${item.role === 'student' ? 'selected' : ''}>student</option>
                        <option value="assistant" ${item.role === 'assistant' ? 'selected' : ''}>assistant</option>
                      </select>
                    </td>
                    <td>
                      <select class="select-input class-member-status" data-user-id="${item.user_id}">
                        <option value="active" ${item.status === 'active' ? 'selected' : ''}>active</option>
                        <option value="removed" ${item.status === 'removed' ? 'selected' : ''}>removed</option>
                      </select>
                    </td>
                    <td><button class="ghost-button class-member-save" type="button" data-user-id="${item.user_id}">Save</button></td>
                  </tr>
                `).join('')}
              </tbody>
            </table>
          ` : renderTeachingEmpty('No members matched the current filter.')}
        </aside>
      </section>
      <section class="detail-grid" style="margin-top:18px; align-items:start;">
        <article class="detail-card">
          <div class="view-header compact">
            <div>
              <h3>Assignment Performance</h3>
              <p class="view-subtitle">Completion and progress across all assignments in this class.</p>
            </div>
          </div>
          <form id="teacher-assignment-filter-form" class="teaching-filter-bar" style="margin-top:14px; margin-bottom:12px;">
            <select class="select-input" name="assignment_status">
              <option value="all" ${assignmentStatusFilter === 'all' ? 'selected' : ''}>all status</option>
              <option value="upcoming" ${assignmentStatusFilter === 'upcoming' ? 'selected' : ''}>upcoming</option>
              <option value="running" ${assignmentStatusFilter === 'running' ? 'selected' : ''}>running</option>
              <option value="ended" ${assignmentStatusFilter === 'ended' ? 'selected' : ''}>ended</option>
            </select>
            <button class="ghost-button" type="submit">Apply</button>
          </form>
          ${filteredAssignments.length ? `
            <table class="data-table assignment-analytics-table">
              <thead><tr><th>Assignment</th><th>Type</th><th>Status</th><th>Problems</th><th>Started</th><th>Completed</th><th>Rate</th><th>Actions</th></tr></thead>
              <tbody>
                ${filteredAssignments.map((item) => `
                  <tr>
                    <td><strong>${escapeHTML(item.title)}</strong></td>
                    <td><span class="status-pill ${teachingStatusClass(item.type)}">${escapeHTML(item.type)}</span></td>
                    <td><span class="status-pill ${teachingStatusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                    <td>${item.problem_count}</td>
                    <td>${item.started_count}</td>
                    <td>${item.completed_count}</td>
                    <td><span class="status-pill ${item.completion_rate >= 60 ? 'status-accepted' : item.completion_rate > 0 ? 'status-pending' : 'status-neutral'}">${item.completion_rate}%</span></td>
                    <td>
                      <a class="table-link" href="#/teacher/assignments/${item.assignment_id}">Teacher View</a>
                      <span class="table-separator">/</span>
                      <a class="table-link" href="#/assignments/${item.assignment_id}">Student View</a>
                    </td>
                  </tr>
                `).join('')}
              </tbody>
            </table>
          ` : renderTeachingEmpty('No assignments matched the current filter.')}
        </article>
        <aside class="detail-card">
          <div class="view-header compact">
            <div>
              <h3>Top Students</h3>
              <p class="view-subtitle">By solved problems and completed assignments.</p>
            </div>
          </div>
          ${analytics.top_students?.length ? `
            <div class="teaching-stack">
              ${analytics.top_students.map((item, index) => `
                <article class="teaching-ranking-card">
                  <div class="view-header compact">
                    <div>
                      <h4 class="teaching-card-title">#${index + 1} ${escapeHTML(item.username)}</h4>
                      <p class="view-subtitle mono">${escapeHTML(item.userid || '-')} / ${escapeHTML(item.role)}</p>
                    </div>
                    <span class="status-pill ${item.completed_assignments > 0 ? 'status-accepted' : item.solved_problem_count > 0 ? 'status-pending' : 'status-neutral'}">${item.solved_problem_count} solved</span>
                  </div>
                  ${renderTeachingMetaList([
                    { label: 'Completed Assignments', value: String(item.completed_assignments) },
                    { label: 'Last Submission', value: item.last_submission_at || '-', mono: true },
                  ])}
                </article>
              `).join('')}
            </div>
          ` : renderTeachingEmpty('No student activity yet.')}
        </aside>
      </section>
    `;

    document.getElementById('teacher-class-edit-form')?.addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch(`/teacher/classes/${id}`, {
          method: 'PUT',
          body: JSON.stringify({
            name: form.get('name'),
            description: form.get('description'),
            status: form.get('status'),
          }),
        });
        setFlash('Class updated', false);
        return renderTeacherClassDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById('teacher-assignment-form')?.addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch(`/teacher/classes/${id}/assignments`, {
          method: 'POST',
          body: JSON.stringify({
            title: form.get('title'),
            type: form.get('type'),
            playlist_id: Number(form.get('playlist_id')),
            description: form.get('description'),
            start_at: toRFC3339(form.get('start_at')),
            due_at: toRFC3339(form.get('due_at')),
          }),
        });
        setFlash('Assignment created', false);
        return renderTeacherClassDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById('teacher-member-filter-form')?.addEventListener('submit', (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const params = new URLSearchParams();
      const nextMemberKeyword = String(form.get('member_keyword') || '').trim();
      const nextMemberStatus = String(form.get('member_status') || 'all').trim();
      if (nextMemberKeyword) params.set('member_keyword', nextMemberKeyword);
      if (nextMemberStatus && nextMemberStatus !== 'all') params.set('member_status', nextMemberStatus);
      if (assignmentStatusFilter && assignmentStatusFilter !== 'all') params.set('assignment_status', assignmentStatusFilter);
      location.hash = `#/teacher/classes/${id}${params.toString() ? `?${params.toString()}` : ''}`;
    });

    document.getElementById('teacher-assignment-filter-form')?.addEventListener('submit', (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const params = new URLSearchParams();
      const nextAssignmentStatus = String(form.get('assignment_status') || 'all').trim();
      if (memberKeyword) params.set('member_keyword', memberKeyword);
      if (memberStatusFilter && memberStatusFilter !== 'all') params.set('member_status', memberStatusFilter);
      if (nextAssignmentStatus && nextAssignmentStatus !== 'all') params.set('assignment_status', nextAssignmentStatus);
      location.hash = `#/teacher/classes/${id}${params.toString() ? `?${params.toString()}` : ''}`;
    });

    document.querySelectorAll('.class-member-save').forEach((button) => {
      button.addEventListener('click', async () => {
        const userID = button.dataset.userId;
        const role = document.querySelector(`.class-member-role[data-user-id="${userID}"]`)?.value || 'student';
        const status = document.querySelector(`.class-member-status[data-user-id="${userID}"]`)?.value || 'active';
        try {
          await apiFetch(`/teacher/classes/${id}/members/${userID}`, {
            method: 'PUT',
            body: JSON.stringify({ role, status }),
          });
          setFlash(`Updated member #${userID}`, false);
          return renderTeacherClassDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load teacher class detail failed: ${escapeHTML(err.message)}</p></div>`;
  }
}


async function renderTeacherAssignmentOverview(id) {
  if (!isTeacherUser()) {
    app.innerHTML = `<div class="detail-card"><p>Teacher permission required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading assignment overview...</p></div>`;
  try {
    const [detail, playlistsResp] = await Promise.all([
      apiFetch(`/teacher/assignments/${id}`, { method: 'GET' }),
      apiFetch(`/teacher/playlists?page=1&page_size=100`, { method: 'GET' }),
    ]);
    const playlists = playlistsResp.list || [];
    const members = detail.members || [];
    const query = getHashQueryParams();
    const memberKeyword = (query.get('member_keyword') || '').trim().toLowerCase();
    const progressFilter = (query.get('progress_status') || 'all').trim().toLowerCase();
    const filteredMembers = members.filter((item) => {
      const haystack = `${item.username || ''} ${item.userid || ''}`.toLowerCase();
      const matchesKeyword = !memberKeyword || haystack.includes(memberKeyword);
      const status = (item.progress_status || '').toLowerCase();
      const matchesProgress = progressFilter === 'all' || status === progressFilter;
      return matchesKeyword && matchesProgress;
    });
    const completedFiltered = filteredMembers.filter((item) => (item.progress_status || '').toLowerCase() === 'completed').length;
    const startedFiltered = filteredMembers.filter((item) => {
      const status = (item.progress_status || '').toLowerCase();
      return status === 'in_progress' || status === 'completed';
    }).length;
    const completionRate = filteredMembers.length ? Math.round((completedFiltered / filteredMembers.length) * 100) : 0;
    const problemHeaders = (members[0]?.problem_statuses || []).map((item) => `
      <th title="${escapeHTML(item.title)}">${escapeHTML(item.display_id || String(item.display_order))}</th>
    `).join('');

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">${escapeHTML(detail.class_name)} / ${escapeHTML(detail.playlist_title)} / teacher overview</p>
        </div>
        <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
          <span class="status-pill ${teachingStatusClass(detail.type)}">${escapeHTML(detail.type)}</span>
          <span class="status-pill ${teachingStatusClass(detail.status)}">${escapeHTML(detail.status)}</span>
          <a class="ghost-button" href="#/teacher/classes/${detail.class_id}">Back to Class</a>
          <a class="ghost-button" href="#/assignments/${detail.id}">Student View</a>
        </div>
      </div>
      ${renderTeachingMetaList([
        { label: 'Students', value: String(detail.member_count) },
        { label: 'Completed', value: String(detail.completed_count) },
        { label: 'Started', value: String(detail.started_count) },
        { label: 'Problems', value: String(detail.problem_count) },
        { label: 'Window', value: `${detail.start_at || '-'} -> ${detail.due_at || '-'}`, mono: true },
      ])}
      <section class="detail-grid" style="margin-top:18px; align-items:start;">
        <article class="detail-card">
          <h3>Edit Assignment</h3>
          <form id="teacher-assignment-edit-form">
            <label class="field-label">Title</label>
            <input class="text-input" name="title" value="${escapeHTML(detail.title)}" required />
            <label class="field-label">Type</label>
            <select class="select-input" name="type">
              <option value="homework" ${detail.type === 'homework' ? 'selected' : ''}>homework</option>
              <option value="exam" ${detail.type === 'exam' ? 'selected' : ''}>exam</option>
            </select>
            <label class="field-label">Playlist</label>
            <select class="select-input" name="playlist_id">
              ${playlists.map((item) => `<option value="${item.id}" ${item.id === detail.playlist_id ? 'selected' : ''}>#${item.id} · ${escapeHTML(item.title)}</option>`).join('')}
            </select>
            <label class="field-label">Description</label>
            <textarea class="text-area" name="description" rows="6">${escapeHTML(detail.description || '')}</textarea>
            <label class="field-label">Start At</label>
            <input class="text-input" type="datetime-local" name="start_at" value="${toDatetimeLocalValue(detail.start_at)}" />
            <label class="field-label">Due At</label>
            <input class="text-input" type="datetime-local" name="due_at" value="${toDatetimeLocalValue(detail.due_at)}" />
            <div style="margin-top:14px; display:flex; gap:10px; flex-wrap:wrap;">
              <button class="primary-button" type="submit">Save Assignment</button>
              <button class="ghost-button" id="teacher-assignment-delete-button" type="button">Delete Assignment</button>
            </div>
          </form>
        </article>
        <aside class="detail-card">
          <div class="view-header compact">
            <div>
              <h3>Progress Summary</h3>
              <p class="view-subtitle">Filter students by keyword and progress state.</p>
            </div>
            <span class="status-pill status-neutral">${filteredMembers.length} / ${detail.member_count}</span>
          </div>
          <form id="teacher-assignment-member-filter-form" class="teaching-filter-bar" style="margin-top:14px;">
            <input class="text-input" name="member_keyword" placeholder="Search username or userid" value="${escapeHTML(query.get('member_keyword') || '')}" />
            <select class="select-input" name="progress_status">
              <option value="all" ${progressFilter === 'all' ? 'selected' : ''}>all progress</option>
              <option value="not_started" ${progressFilter === 'not_started' ? 'selected' : ''}>not_started</option>
              <option value="in_progress" ${progressFilter === 'in_progress' ? 'selected' : ''}>in_progress</option>
              <option value="completed" ${progressFilter === 'completed' ? 'selected' : ''}>completed</option>
            </select>
            <button class="ghost-button" type="submit">Apply</button>
            <a class="ghost-button" href="#/teacher/assignments/${id}">Clear</a>
          </form>
          <div class="verdict-summary-grid" style="margin-top:14px;">
            <div class="verdict-summary-card"><span class="status-pill status-neutral">Students</span><strong>${filteredMembers.length}</strong></div>
            <div class="verdict-summary-card"><span class="status-pill status-accepted">Completed</span><strong>${completedFiltered}</strong></div>
            <div class="verdict-summary-card"><span class="status-pill status-pending">Started</span><strong>${startedFiltered}</strong></div>
            <div class="verdict-summary-card"><span class="status-pill status-neutral">Completion Rate</span><strong>${completionRate}%</strong></div>
          </div>
          <div style="margin-top:14px;">
            ${filteredMembers.length ? `
              <div class="teaching-stack">
                ${filteredMembers.map((item) => `
                  <article class="teaching-list-card">
                    <div class="view-header compact">
                      <div>
                        <h4 class="teaching-card-title">${escapeHTML(item.username)}</h4>
                        <p class="view-subtitle mono">${escapeHTML(item.userid || '-')}</p>
                      </div>
                      <span class="status-pill ${teachingStatusClass(item.progress_status)}">${escapeHTML(item.progress_status)}</span>
                    </div>
                    ${renderTeachingMetaList([
                      { label: 'Role', value: item.role },
                      { label: 'Solved', value: `${item.solved_count}/${item.total_count}` },
                      { label: 'Last Submission', value: item.last_submission_at || '-', mono: true },
                    ])}
                  </article>
                `).join('')}
              </div>
            ` : renderTeachingEmpty('No active class member yet.')}
          </div>
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px; overflow:auto;">
        <div class="view-header compact">
          <div>
            <h3>Per Student Problem Matrix</h3>
            <p class="view-subtitle">Accepted problems are highlighted; other cells show the latest verdict.</p>
          </div>
        </div>
        ${filteredMembers.length ? `
          <table class="data-table assignment-overview-table">
            <thead>
              <tr>
                <th>Student</th>
                <th>Role</th>
                <th>Progress</th>
                <th>Last Submission</th>
                ${problemHeaders}
              </tr>
            </thead>
            <tbody>
              ${filteredMembers.map((member) => `
                <tr>
                  <td>
                    <strong>${escapeHTML(member.username)}</strong>
                    <div class="view-subtitle mono">${escapeHTML(member.userid || '-')}</div>
                  </td>
                  <td><span class="status-pill ${teachingStatusClass(member.role)}">${escapeHTML(member.role)}</span></td>
                  <td>${renderAssignmentProgressPill(member)}</td>
                  <td class="mono">${escapeHTML(member.last_submission_at || '-')}</td>
                  ${(member.problem_statuses || []).map((problem) => `
                    <td>
                      ${problem.solved
                        ? '<span class="status-pill status-accepted">AC</span>'
                        : renderProblemStatusPill(problem.last_status || null)}
                    </td>
                  `).join('')}
                </tr>
              `).join('')}
            </tbody>
          </table>
        ` : renderTeachingEmpty('No student matched the current filter.')}
      </section>
    `;

    document.getElementById('teacher-assignment-edit-form')?.addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch(`/teacher/assignments/${id}`, {
          method: 'PUT',
          body: JSON.stringify({
            title: form.get('title'),
            type: form.get('type'),
            playlist_id: Number(form.get('playlist_id')),
            description: form.get('description'),
            start_at: toRFC3339(form.get('start_at')),
            due_at: toRFC3339(form.get('due_at')),
          }),
        });
        setFlash('Assignment updated', false);
        return renderTeacherAssignmentOverview(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById('teacher-assignment-member-filter-form')?.addEventListener('submit', (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const params = new URLSearchParams();
      const nextKeyword = String(form.get('member_keyword') || '').trim();
      const nextProgress = String(form.get('progress_status') || 'all').trim();
      if (nextKeyword) params.set('member_keyword', nextKeyword);
      if (nextProgress && nextProgress !== 'all') params.set('progress_status', nextProgress);
      location.hash = `#/teacher/assignments/${id}${params.toString() ? `?${params.toString()}` : ''}`;
    });

    document.getElementById('teacher-assignment-delete-button')?.addEventListener('click', async () => {
      if (!window.confirm(`Delete assignment #${id}?`)) return;
      try {
        await apiFetch(`/teacher/assignments/${id}`, { method: 'DELETE' });
        setFlash('Assignment deleted', false);
        location.hash = `#/teacher/classes/${detail.class_id}`;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load teacher assignment overview failed: ${escapeHTML(err.message)}</p></div>`;
  }
}
