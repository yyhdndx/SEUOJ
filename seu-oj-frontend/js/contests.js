// Contest domain pages and helpers
async function renderContests() {
  const queryString = getCurrentHashPath().split("?")[1] || "";
  const params = new URLSearchParams(queryString);
  const page = params.get("page") || "1";
  const pageSize = params.get("page_size") || "20";
  const keyword = params.get("keyword") || "";
  const status = params.get("status") || "";

  app.innerHTML = `<div class="detail-card"><p>Loading contests...</p></div>`;
  try {
    const query = new URLSearchParams({ page, page_size: pageSize });
    if (keyword) query.set("keyword", keyword);
    if (status) query.set("status", status);
    const data = await apiFetch(`/contests?${query.toString()}`, { method: "GET" });

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Contests</h1>
          <p class="view-subtitle">Upcoming, running, and archived contests built on the current problem and submission system.</p>
        </div>
      </div>
      <form id="contest-filters" class="toolbar">
        <div>
          <label class="field-label">Keyword</label>
          <input class="text-input" name="keyword" value="${escapeHTML(keyword)}" placeholder="search title" />
        </div>
        <div>
          <label class="field-label">Status</label>
          <select class="select-input" name="status">
            <option value="">All</option>
            ${["upcoming", "running", "ended"].map((item) => `<option value="${item}" ${status === item ? "selected" : ""}>${item}</option>`).join("")}
          </select>
        </div>
        <div>
          <label class="field-label">Page</label>
          <input class="text-input" name="page" type="number" min="1" value="${escapeHTML(page)}" />
        </div>
        <div>
          <label class="field-label">Page Size</label>
          <input class="text-input" name="page_size" type="number" min="1" max="100" value="${escapeHTML(pageSize)}" />
        </div>
        <div class="full" style="display:flex; gap:10px; align-items:center;">
          <button class="ghost-button" type="submit">Apply Filters</button>
          <button class="ghost-button" type="button" id="clear-contest-filters">Clear</button>
        </div>
      </form>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Title</th>
            <th>Status</th>
            <th>Window</th>
            <th>Hint</th>
            <th>Problems</th>
            <th>Registrations</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          ${(data.list || []).map((item) => `
            <tr>
              <td>${item.id}</td>
              <td>
                <div style="display:flex; flex-direction:column; gap:4px;">
                  <strong>${escapeHTML(item.title)}</strong>
                  <span class="mono">${escapeHTML(item.rule_type)}</span>
                </div>
              </td>
              <td><span class="status-pill ${contestStatusClass(item.status)}">${escapeHTML(item.status)}</span></td>
              <td class="mono">${escapeHTML(item.start_time)}<br>${escapeHTML(item.end_time)}</td>
              <td class="mono">${renderContestStatusHint(item)}</td>
              <td>${item.problem_count}</td>
              <td>${item.registered_count}</td>
              <td><a class="table-link" href="#/contests/${item.id}">Open</a></td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    `;

    document.getElementById("contest-filters").addEventListener("submit", (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const next = new URLSearchParams();
      const nextKeyword = (form.get("keyword") || "").toString().trim();
      const nextStatus = (form.get("status") || "").toString().trim();
      const nextPage = (form.get("page") || "1").toString().trim() || "1";
      const nextPageSize = (form.get("page_size") || "20").toString().trim() || "20";
      next.set("page", nextPage);
      next.set("page_size", nextPageSize);
      if (nextKeyword) next.set("keyword", nextKeyword);
      if (nextStatus) next.set("status", nextStatus);
      location.hash = `#/contests?${next.toString()}`;
    });
    document.getElementById("clear-contest-filters").addEventListener("click", () => {
      location.hash = "#/contests";
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load contests failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderContestDetail(id) {
  app.innerHTML = `<div class="detail-card"><p>Loading contest...</p></div>`;
  try {
    const [contest, ranklist, announcements] = await Promise.all([
      apiFetch(`/contests/${id}`, { method: "GET" }),
      apiFetch(`/contests/${id}/ranklist`, { method: "GET" }).catch(() => null),
      apiFetch(`/contests/${id}/announcements?page=1&page_size=20`, { method: "GET" }).catch(() => ({ list: [] })),
    ]);

    let me = null;
    let problems = null;
    let contestSubmissions = [];
    if (state.token) {
      me = await apiFetch(`/contests/${id}/me`, { method: "GET" }).catch(() => null);
      if (me?.can_view_problems) {
        problems = await apiFetch(`/contests/${id}/problems`, { method: "GET" }).catch(() => null);
      }
      contestSubmissions = await apiFetch(`/submissions/my?page=1&page_size=5&contest_id=${id}`, { method: "GET" }).then((data) => data.list || []).catch(() => []);
    }

    const problemSetHint = me?.can_view_problems
      ? (me?.can_submit
          ? "Contest problems are available for solving."
          : me?.practice_enabled
            ? "Contest has ended. Practice submissions remain open."
            : "Contest problems are viewable, but submissions are currently closed.")
      : (state.token ? "Register and wait for the contest to start before opening problems." : "Login and register to access contest problems.");
    const ranklistHint = ranklist?.ranklist_frozen
      ? `Ranklist frozen at ${escapeHTML(ranklist.ranklist_freeze_at || "-")}. Cells marked with * hide post-freeze submissions until contest ends.`
      : "Live ACM-style standings computed from contest-tagged submissions.";

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(contest.title)}</h1>
          <p class="view-subtitle">${escapeHTML(contest.rule_type)} contest - ${escapeHTML(contest.status)} - ${renderContestStatusHint(contest)}</p>
        </div>
        <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
          <span class="status-pill ${contestStatusClass(contest.status)}">${escapeHTML(contest.status)}</span>
          ${contest.allow_practice ? `<span class="status-pill status-neutral">practice</span>` : ""}
          ${contest.ranklist_frozen ? `<span class="status-pill status-pending">frozen</span>` : contest.ranklist_freeze_at ? `<span class="status-pill status-neutral">freeze set</span>` : ""}
          <a class="ghost-button" href="#/forum?scope_type=contest&scope_id=${encodeURIComponent(contest.id)}">Discuss</a>
          ${state.token && me?.can_register ? `<button class="primary-button" id="contest-register-btn">Register</button>` : ""}
          ${me?.can_view_problems ? `<a class="ghost-button" href="#contest-problems">Problems</a>` : ""}
        </div>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          <h3>Description</h3>
          ${renderMarkdownBlock(contest.description || "No description.")}
        </article>
        <aside class="detail-card">
          <h3>Overview</h3>
          <div class="metric-list">
            <div class="metric"><span class="metric-label">Rule</span><span class="metric-value mono">${escapeHTML(contest.rule_type)}</span></div>
            <div class="metric"><span class="metric-label">Start</span><span class="metric-value mono">${escapeHTML(contest.start_time)}</span></div>
            <div class="metric"><span class="metric-label">End</span><span class="metric-value mono">${escapeHTML(contest.end_time)}</span></div>
            <div class="metric"><span class="metric-label">Practice</span><span class="metric-value">${contest.allow_practice ? "Enabled" : "Closed"}</span></div>
            <div class="metric"><span class="metric-label">Freeze At</span><span class="metric-value mono">${escapeHTML(contest.ranklist_freeze_at || "-")}</span></div>
            <div class="metric"><span class="metric-label">Ranklist</span><span class="metric-value">${contest.ranklist_frozen ? "Frozen" : "Live"}</span></div>
            <div class="metric"><span class="metric-label">Problems</span><span class="metric-value">${contest.problem_count}</span></div>
            <div class="metric"><span class="metric-label">Registrations</span><span class="metric-value">${contest.registered_count}</span></div>
            ${me ? `<div class="metric"><span class="metric-label">Registered</span><span class="metric-value">${me.registered ? "Yes" : "No"}</span></div>` : ""}
          </div>
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Contest Announcements</h3>
            <p class="view-subtitle">Pinned updates and clarifications released for this contest.</p>
          </div>
        </div>
        ${renderContestAnnouncementList(announcements.list || [], { admin: false, contestID: Number(id) })}
      </section>
      <section class="detail-card" id="contest-problems" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Problem Set</h3>
            <p class="view-subtitle">${problemSetHint}</p>
          </div>
        </div>
        ${problems ? renderContestProblemTable(id, problems.list || []) : `<p class="view-subtitle">Problem set is currently unavailable for your account.</p>`}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>My Contest Submissions</h3>
            <p class="view-subtitle">Recent submissions tagged with this contest.</p>
          </div>
          <div style="display:flex; gap:10px; align-items:center;"><span class="mono">${contestSubmissions.filter((item) => isSubmissionPollingStatus(item.status)).length ? "judging active" : "judging idle"}</span><a class="ghost-button" href="#/submissions?contest_id=${id}">Open Filtered List</a></div>
        </div>
        ${renderContestRecentSubmissions(contestSubmissions)}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Standings</h3>
            <p class="view-subtitle">${ranklistHint}</p>
          </div>
          ${ranklist?.ranklist_frozen ? `<span class="status-pill status-pending">Frozen</span>` : `<span class="status-pill status-accepted">Live</span>`}
        </div>
        ${ranklist ? renderContestRanklistTable(ranklist) : `<p class="view-subtitle">Standings unavailable.</p>`}
      </section>
    `;

    startContestPolling(id, contest.status);

    const registerButton = document.getElementById("contest-register-btn");
    if (registerButton) {
      registerButton.addEventListener("click", async () => {
        try {
          await apiFetch(`/contests/${id}/register`, { method: "POST" });
          setFlash("Contest registration completed", false);
          await renderContestDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    }
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load contest failed: ${escapeHTML(err.message)}</p></div>`;
  }
}
async function renderContestProblemDetail(contestID, problemID) {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before opening contest problems.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading contest problem...</p></div>`;
  try {
    const [contest, problem] = await Promise.all([
      apiFetch(`/contests/${contestID}`, { method: "GET" }),
      apiFetch(`/contests/${contestID}/problems/${problemID}`, { method: "GET" }),
    ]);
    state.problemDetail = problem;
    if (!state.runResult || state.runResult.problemID !== problem.id || state.runResult.contestID !== Number(contestID)) {
      state.runResult = null;
    }
    const draft = readSubmissionDraft(problem.id);
    const selectedLanguage = draft?.language || "cpp";
    const initialCode = draft?.code || getDefaultCodeTemplate(selectedLanguage);
    const sampleCases = Array.isArray(problem.testcases)
      ? problem.testcases.filter((item) => item.case_type === "sample")
      : [];
    const recentSubmissions = await loadRecentProblemSubmissions(problem.id, contestID);

    app.innerHTML = `
      <section class="problem-workbench resizable" id="problem-workbench" style="--problem-pane-width:${state.workbenchLeftWidth}%;">
        <article class="problem-pane">
          <div class="pane-header">
            <div>
              <h2 class="pane-title">${escapeHTML(problem.title)}</h2>
              <p class="view-subtitle">Contest #${contest.id} / ${escapeHTML(contest.title)} / ${escapeHTML(problem.judge_mode)}</p>
            </div>
            <div class="pill-row">
              <a class="ghost-button" href="#/contests/${contestID}">Back to Contest</a>
            </div>
          </div>
          <div class="pane-content">
            ${renderProblemBlock("Description", problem.description)}
            ${renderProblemBlock("Input", problem.input_desc)}
            ${renderProblemBlock("Output", problem.output_desc)}
            ${renderProblemCodeBlock("Sample Input", problem.sample_input)}
            ${renderProblemCodeBlock("Sample Output", problem.sample_output)}
            ${sampleCases.length ? `
              <div class="detail-block">
                <h3>Sample Testcases</h3>
                <p class="view-subtitle">${sampleCases.length} sample case(s) are available.</p>
                ${renderSampleCaseList(sampleCases)}
              </div>
            ` : ""}
            ${renderProblemBlock("Source / Hint", `${problem.source || ""}\n${problem.hint || ""}`.trim())}
            ${renderProblemRecentSubmissions(problem.id, recentSubmissions, contestID)}
          </div>
        </article>
        <div class="workbench-divider" id="workbench-divider" aria-hidden="true"></div>
        <aside class="editor-pane">
          <div class="pane-content">
          <form id="submit-form" class="editor-form">
            <div class="editor-surface">
              <textarea class="text-area" id="problem-code-editor" name="code" placeholder="#include <iostream>...">${escapeHTML(initialCode)}</textarea>
            </div>
            <div class="editor-bottom-stack">
              <div class="editor-submit-strip">
                <div class="submit-language-group">
                  <label class="submit-language-label" for="problem-language-select">Language</label>
                  <select class="select-input submit-language-select" name="language" id="problem-language-select">
                    ${submissionLanguageOptions.map(([value, label]) => `
                      <option value="${value}" ${selectedLanguage === value ? "selected" : ""}>${escapeHTML(label)}</option>
                    `).join("")}
                  </select>
                </div>
                <div class="submit-action-group">
                  <button class="ghost-button submit-compact-button" type="button" id="run-sample-btn">Run</button>
                  <button class="primary-button submit-compact-button" type="submit">Submit</button>
                </div>
              </div>
              ${renderRunResultPanel()}
            </div>
          </form>
          </div>
        </aside>
      </section>
    `;

    initProblemWorkbenchUI();
    initRunResultUI();

    const codeEditor = document.getElementById("problem-code-editor");
    const languageSelect = document.getElementById("problem-language-select");
    let currentLanguage = selectedLanguage;
    languageSelect?.addEventListener("change", (event) => {
      const nextLanguage = event.currentTarget.value;
      const previousTemplate = getDefaultCodeTemplate(currentLanguage);
      if (!codeEditor.value.trim() || codeEditor.value === previousTemplate) {
        codeEditor.value = getDefaultCodeTemplate(nextLanguage);
      }
      currentLanguage = nextLanguage;
    });

    document.getElementById("run-sample-btn").addEventListener("click", async () => {
      const form = new FormData(document.getElementById("submit-form"));
      const code = (form.get("code") || "").toString();
      const language = (form.get("language") || "cpp").toString();
      saveSubmissionDraft(problem.id, language, code);
      try {
        const result = await apiFetch("/submissions/run", {
          method: "POST",
          body: JSON.stringify({
            problem_id: Number(problem.id),
            contest_id: Number(contestID),
            language,
            code,
          }),
        });
        state.runResult = { problemID: problem.id, contestID: Number(contestID), ...result };
        renderContestProblemDetail(contestID, problemID);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("submit-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const language = (form.get("language") || "").toString();
      const code = (form.get("code") || "").toString();
      saveSubmissionDraft(problem.id, language, code);
      try {
        const result = await apiFetch("/submissions", {
          method: "POST",
          body: JSON.stringify({
            problem_id: Number(problem.id),
            contest_id: Number(contestID),
            language,
            code,
          }),
        });
        setFlash(`Contest submission created: #${result.submission_id}`, false);
        location.hash = `#/submissions/${result.submission_id}`;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load contest problem failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAdminContests() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before managing contests.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access contest management.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading contests...</p></div>`;
  try {
    const data = await apiFetch("/admin/contests?page=1&page_size=100", { method: "GET" });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Manage Contests</h1>
          <p class="view-subtitle">Create and maintain contest schedules, problem sets, and visibility.</p>
        </div>
        <div>
          <a class="primary-button" href="#/admin/contests/new">New Contest</a>
        </div>
      </div>
      <section class="detail-card">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Title</th>
              <th>Status</th>
              <th>Window</th>
              <th>Problems</th>
              <th>Registrations</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            ${(data.list || []).map((item) => `
              <tr>
                <td>${item.id}</td>
                <td>${escapeHTML(item.title)}</td>
                <td><span class="status-pill ${contestStatusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                <td class="mono">${escapeHTML(item.start_time)}<br>${escapeHTML(item.end_time)}</td>
                <td>${item.problem_count}</td>
                <td>${item.registered_count}</td>
                <td>
                  <a class="table-link" href="#/admin/contests/${item.id}">View</a><span class="mono"> | </span><a class="table-link" href="#/admin/contests/${item.id}/edit">Edit</a>
                  <span class="mono"> | </span>
                  <button class="link-button admin-contest-delete" data-contest-id="${item.id}">Delete</button>
                </td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      </section>
    `;

    document.querySelectorAll(".admin-contest-delete").forEach((button) => {
      button.addEventListener("click", async () => {
        const id = button.dataset.contestId;
        try {
          await apiFetch(`/admin/contests/${id}`, { method: "DELETE" });
          setFlash(`Contest #${id} deleted`, false);
          await renderAdminContests();
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load admin contests failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAdminContestDetail(id) {
  if (!state.token || state.user?.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access contest detail.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading contest detail...</p></div>`;
  try {
    const [detail, ranklist, announcements] = await Promise.all([
      apiFetch(`/admin/contests/${id}`, { method: "GET" }),
      apiFetch(`/admin/contests/${id}/ranklist`, { method: "GET" }).catch(() => null),
      apiFetch(`/admin/contests/${id}/announcements?page=1&page_size=50`, { method: "GET" }).catch(() => ({ list: [] })),
    ]);
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">${escapeHTML(detail.rule_type)} contest - ${escapeHTML(detail.status)} - created by #${detail.created_by}</p>
        </div>
        <div style="display:flex; gap:10px; flex-wrap:wrap;">
          <a class="ghost-button" href="#/admin/contests">Back</a>
          <a class="ghost-button" href="#/admin/contests/${detail.id}/announcements/new">New Announcement</a>
          <a class="primary-button" href="#/admin/contests/${detail.id}/edit">Edit Contest</a>
        </div>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          <h3>Description</h3>
          ${renderMarkdownBlock(detail.description || "No description.")}
        </article>
        <aside class="detail-card">
          <h3>Overview</h3>
          <div class="metric-list">
            <div class="metric"><span class="metric-label">Public</span><span class="metric-value">${detail.is_public ? "Yes" : "No"}</span></div>
            <div class="metric"><span class="metric-label">Practice</span><span class="metric-value">${detail.allow_practice ? "Enabled" : "Closed"}</span></div>
            <div class="metric"><span class="metric-label">Freeze At</span><span class="metric-value mono">${escapeHTML(detail.ranklist_freeze_at || "-")}</span></div>
            <div class="metric"><span class="metric-label">Problems</span><span class="metric-value">${detail.problem_count}</span></div>
            <div class="metric"><span class="metric-label">Registrations</span><span class="metric-value">${detail.registered_count}</span></div>
            <div class="metric"><span class="metric-label">Start</span><span class="metric-value mono">${escapeHTML(detail.start_time)}</span></div>
            <div class="metric"><span class="metric-label">End</span><span class="metric-value mono">${escapeHTML(detail.end_time)}</span></div>
            <div class="metric"><span class="metric-label">Updated</span><span class="metric-value mono">${escapeHTML(detail.updated_at)}</span></div>
          </div>
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Contest Announcements</h3>
            <p class="view-subtitle">Scoped notices for contestants. Manage them here.</p>
          </div>
        </div>
        ${renderContestAnnouncementList(announcements.list || [], { admin: true, contestID: detail.id })}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Contest Problems</h3>
            <p class="view-subtitle">Internal contest set order, codes, and linked problems.</p>
          </div>
        </div>
        ${detail.problems?.length ? renderContestProblemTable(detail.id, detail.problems) : `<p class="view-subtitle">No contest problems configured.</p>`}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Admin Ranklist</h3>
            <p class="view-subtitle">Full standings view. Freeze is ignored for admins.</p>
          </div>
        </div>
        ${ranklist ? renderContestRanklistTable(ranklist) : `<p class="view-subtitle">Standings unavailable.</p>`}
      </section>
    `;
    document.querySelectorAll(".contest-announcement-delete").forEach((button) => {
      button.addEventListener("click", async () => {
        const announcementID = button.dataset.announcementId;
        try {
          await apiFetch(`/admin/contests/${id}/announcements/${announcementID}`, { method: "DELETE" });
          setFlash("Contest announcement deleted", false);
          renderAdminContestDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load admin contest detail failed: ${escapeHTML(err.message)}</p></div>`;
  }
}
function renderAdminContestCreate() {
  return renderAdminContestForm(null, null);
}

async function renderAdminContestEdit(id) {
  if (!state.token || state.user?.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can edit contests.</p></div>`;
    return;
  }
  // 参数校验：id 必须为正整数，否则跳转到新建比赛页面
  if (!/^[0-9]+$/.test(String(id))) {
    location.hash = "#/admin/contests/new";
    return;
  }
  try {
    const detail = await apiFetch(`/admin/contests/${id}`, { method: "GET" });
    return renderAdminContestForm(id, detail);
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load contest failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function loadAdminContestProblems() {
  const pageSize = 100;
  const firstPage = await apiFetch(`/admin/problems?page=1&page_size=${pageSize}&include_hidden=true`, { method: "GET" });
  const firstList = firstPage.list || [];
  const total = Math.max(Number(firstPage.total ?? 0) || 0, firstList.length);
  const totalPages = Math.ceil(total / pageSize);
  if (totalPages <= 1) {
    return firstList;
  }

  const remainingPages = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, index) => {
      const page = index + 2;
      return apiFetch(`/admin/problems?page=${page}&page_size=${pageSize}&include_hidden=true`, { method: "GET" });
    }),
  );

  return firstList.concat(remainingPages.flatMap((pageData) => pageData.list || []));
}

async function renderAdminContestForm(id, initial) {
  if (!state.token || state.user?.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access contest editing.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading contest form...</p></div>`;
  try {
    const problems = await loadAdminContestProblems();
    const initialProblems = (initial?.problems && initial.problems.length)
      ? initial.problems
      : [{ problem_id: problems[0]?.id || 0, problem_code: "A", display_order: 1 }];

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${id ? "Edit Contest" : "New Contest"}</h1>
          <p class="view-subtitle">ACM rule scoreboard, ordered problem set, and open registration.</p>
        </div>
        <div>
          <a class="ghost-button" href="#/admin/contests">Back</a>
        </div>
      </div>
      <form id="contest-form" class="detail-card toolbar">
        <div class="full">
          <label class="field-label">Title</label>
          <input class="text-input" name="title" value="${escapeHTML(initial?.title || "")}" required />
        </div>
        <div class="full">
          <label class="field-label">Description</label>
          <textarea class="text-area" name="description">${escapeHTML(initial?.description || "")}</textarea>
        </div>
        <div>
          <label class="field-label">Rule Type</label>
          <select class="select-input" name="rule_type">
            <option value="acm" selected>acm</option>
          </select>
        </div>
        <div>
          <label class="field-label">Start Time</label>
          <input class="text-input" name="start_time" type="datetime-local" value="${escapeHTML(toDateTimeLocalValue(initial?.start_time))}" required />
        </div>
        <div>
          <label class="field-label">End Time</label>
          <input class="text-input" name="end_time" type="datetime-local" value="${escapeHTML(toDateTimeLocalValue(initial?.end_time))}" required />
        </div>
        <div>
          <label class="field-label">Public</label>
          <input type="checkbox" name="is_public" ${initial?.is_public ?? true ? "checked" : ""} />
        </div>
        <div>
          <label class="field-label">Allow Practice After End</label>
          <input type="checkbox" name="allow_practice" ${initial?.allow_practice ? "checked" : ""} />
        </div>
        <div>
          <label class="field-label">Ranklist Freeze At</label>
          <input class="text-input" name="ranklist_freeze_at" type="datetime-local" value="${escapeHTML(toDateTimeLocalValue(initial?.ranklist_freeze_at))}" />
        </div>
        <div class="full">
          <div class="view-header" style="margin:0 0 10px 0;">
            <div>
              <h3 style="margin:0;">Contest Problems</h3>
              <p class="view-subtitle">Pick existing problems and assign A/B/C style codes.</p>
            </div>
            <button class="ghost-button" type="button" id="add-contest-problem-row">Add Problem</button>
          </div>
          <div id="contest-problem-rows"></div>
        </div>
        <div class="full" style="display:flex; gap:10px; align-items:center;">
          <button class="primary-button" type="submit">${id ? "Update Contest" : "Create Contest"}</button>
          <a class="ghost-button" href="#/admin/contests">Cancel</a>
        </div>
      </form>
    `;

    const rowsContainer = document.getElementById("contest-problem-rows");
    const renderRow = (value = {}) => {
      const wrapper = document.createElement("div");
      wrapper.className = "toolbar";
      wrapper.style.marginBottom = "10px";
      wrapper.innerHTML = `
        <div>
          <label class="field-label">Problem</label>
          <select class="select-input contest-problem-id">
            ${problems.map((problem) => `<option value="${problem.id}" ${Number(value.problem_id) === Number(problem.id) ? "selected" : ""}>#${problem.id} / ${escapeHTML(problem.display_id || "-")} / ${escapeHTML(problem.title)}</option>`).join("")}
          </select>
        </div>
        <div>
          <label class="field-label">Code</label>
          <input class="text-input contest-problem-code" value="${escapeHTML(value.problem_code || "")}" placeholder="A" />
        </div>
        <div>
          <label class="field-label">Order</label>
          <input class="text-input contest-problem-order" type="number" min="1" value="${escapeHTML(value.display_order || "1")}" />
        </div>
        <div style="display:flex; align-items:flex-end;">
          <button class="ghost-button remove-contest-problem-row" type="button">Remove</button>
        </div>
      `;
      wrapper.querySelector(".remove-contest-problem-row").addEventListener("click", () => {
        if (rowsContainer.children.length === 1) {
          return;
        }
        wrapper.remove();
      });
      rowsContainer.appendChild(wrapper);
    };

    initialProblems.forEach((item, index) => {
      renderRow({
        problem_id: item.problem_id,
        problem_code: item.problem_code || String.fromCharCode(65 + index),
        display_order: item.display_order || index + 1,
      });
    });

    document.getElementById("add-contest-problem-row").addEventListener("click", () => {
      renderRow({
        problem_id: problems[0]?.id || 0,
        problem_code: String.fromCharCode(65 + rowsContainer.children.length),
        display_order: rowsContainer.children.length + 1,
      });
    });

    document.getElementById("contest-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const problemRows = Array.from(rowsContainer.children).map((row) => ({
        problem_id: Number(row.querySelector(".contest-problem-id").value),
        problem_code: row.querySelector(".contest-problem-code").value.trim(),
        display_order: Number(row.querySelector(".contest-problem-order").value || 1),
      }));

      try {
        await apiFetch(id ? `/admin/contests/${id}` : "/admin/contests", {
          method: id ? "PUT" : "POST",
          body: JSON.stringify({
            title: form.get("title"),
            description: form.get("description"),
            rule_type: form.get("rule_type"),
            start_time: new Date(form.get("start_time")).toISOString(),
            end_time: new Date(form.get("end_time")).toISOString(),
            is_public: form.get("is_public") === "on",
            allow_practice: form.get("allow_practice") === "on",
            ranklist_freeze_at: form.get("ranklist_freeze_at") ? new Date(form.get("ranklist_freeze_at")).toISOString() : null,
            problems: problemRows,
          }),
        });
        setFlash(`Contest ${id ? "updated" : "created"}`, false);
        location.hash = "#/admin/contests";
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load contest form failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function renderContestProblemTable(contestID, problems) {
  if (!problems.length) {
    return `<p class="view-subtitle">No contest problem is configured yet.</p>`;
  }
  return `
    <table class="data-table">
      <thead>
        <tr>
          <th>Code</th>
          <th>Title</th>
          <th>Judge</th>
          <th>Limits</th>
          <th>Open</th>
        </tr>
      </thead>
      <tbody>
        ${problems.map((item) => `
          <tr>
            <td><span class="status-pill status-neutral">${escapeHTML(item.problem_code)}</span></td>
            <td>${escapeHTML(item.title)}</td>
            <td>${escapeHTML(item.judge_mode)}</td>
            <td>${item.time_limit_ms} ms / ${item.memory_limit_mb} MB</td>
            <td><a class="table-link" href="#/contests/${contestID}/problems/${item.problem_id}">Solve</a></td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;
}

function renderContestRanklistTable(ranklist) {
  const problems = ranklist.problems || [];
  const rows = ranklist.list || [];
  if (!rows.length) {
    return `<p class="view-subtitle">No registered participants or no contest-tagged submissions yet.</p>`;
  }
  return `
    ${ranklist.ranklist_frozen ? `<p class="view-subtitle" style="margin-bottom:12px;">* indicates hidden post-freeze submissions.</p>` : ""}
    <div style="overflow:auto;">
      <table class="data-table">
        <thead>
          <tr>
            <th>Rank</th>
            <th>User</th>
            <th>Solved</th>
            <th>Penalty</th>
            <th>Submissions</th>
            ${problems.map((problem) => `<th>${escapeHTML(problem.problem_code)}</th>`).join("")}
          </tr>
        </thead>
        <tbody>
          ${rows.map((row) => `
            <tr class="${state.user && Number(state.user.id) === Number(row.user_id) ? "row-current-user" : ""}">
              <td>${row.rank}</td>
              <td>${escapeHTML(row.username)}<br><span class="mono">${escapeHTML(row.userid || "")}</span></td>
              <td>${row.solved_count}</td>
              <td>${row.penalty_minutes}</td>
              <td>${row.submission_count}</td>
              ${row.cells.map((cell) => `<td class="${contestRankCellClass(cell)}"${contestRankCellTitle(cell)}>${escapeHTML(contestRankCellLabel(cell))}</td>`).join("")}
            </tr>
          `).join("")}
        </tbody>
      </table>
    </div>
  `;
}

function contestRankCellLabel(cell) {
  let label = ".";
  if (cell.solved) {
    label = cell.wrong_attempts > 0 ? `+${cell.wrong_attempts}` : "+";
  } else if (cell.wrong_attempts > 0) {
    label = `-${cell.wrong_attempts}`;
  } else if (cell.latest_status === "Pending" || cell.latest_status === "Running") {
    label = "?";
  }
  if (cell.frozen_submissions > 0) {
    const suffix = cell.frozen_submissions > 1 ? `*${cell.frozen_submissions}` : "*";
    return label === "." ? suffix : `${label}${suffix}`;
  }
  return label;
}

function contestRankCellClass(cell) {
  const classes = ["mono", "contest-rank-cell"];
  if (cell.solved) {
    classes.push("contest-rank-solved");
  } else if (cell.wrong_attempts > 0) {
    classes.push("contest-rank-tried");
  }
  if (cell.frozen_submissions > 0) {
    classes.push("contest-rank-frozen");
  }
  return classes.join(" ");
}

function contestRankCellTitle(cell) {
  if (!cell.frozen_submissions) {
    return "";
  }
  const count = Number(cell.frozen_submissions) || 0;
  const label = count === 1 ? "1 hidden post-freeze submission" : `${count} hidden post-freeze submissions`;
  return ` title="${escapeHTML(label)}"`;
}

function contestStatusClass(status) {
  if (status === "running") return "status-pending";
  if (status === "ended") return "status-accepted";
  return "status-neutral";
}

function toDateTimeLocalValue(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const offset = date.getTimezoneOffset();
  const adjusted = new Date(date.getTime() - offset * 60000);
  return adjusted.toISOString().slice(0, 16);
}








function renderContestRecentSubmissions(list) {
  if (!list || !list.length) {
    return `<p class="view-subtitle">No contest submission yet.</p>`;
  }
  return `
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Problem</th>
          <th>Mode</th>
          <th>Status</th>
          <th>Passed</th>
          <th>Created</th>
        </tr>
      </thead>
      <tbody>
        ${list.map((item) => `
          <tr>
            <td><a class="table-link" href="#/submissions/${item.id}">${item.id}</a></td>
            <td>${item.problem_id}</td>
            <td>${renderContestModeBadge(item)}</td>
            <td><span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span></td>
            <td>${item.passed_count}/${item.total_count}</td>
            <td class="mono">${escapeHTML(item.created_at)}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;
}

function renderContestModeBadge(item) {
  if (!item?.contest_id) {
    return "-";
  }
  return item.is_practice
    ? `<span class="status-pill status-neutral">practice</span>`
    : `<span class="status-pill status-accepted">official</span>`;
}

function renderContestAnnouncementList(list, options = {}) {
  const admin = !!options.admin;
  const contestID = options.contestID;
  if (!list || !list.length) {
    return `<p class="view-subtitle">No contest announcement yet.</p>`;
  }
  return `
    <div class="announcement-stack">
      ${list.map((item) => `
        <article class="announcement-card">
          <div class="view-header" style="margin-bottom:10px;">
            <div>
              <h3 style="margin:0;">${escapeHTML(item.title)}</h3>
              <p class="view-subtitle">${item.is_pinned ? "Pinned" : "Update"} / ${escapeHTML(item.updated_at || item.created_at)}</p>
            </div>
            <div class="contest-announce-actions">
              ${item.is_pinned ? `<span class="status-pill status-pending">Pinned</span>` : ""}
              ${admin ? `<a class="table-link" href="#/admin/contests/${contestID}/announcements/${item.id}/edit">Edit</a><button class="link-button contest-announcement-delete" data-announcement-id="${item.id}">Delete</button>` : ""}
            </div>
          </div>
          <pre>${escapeHTML(item.content || "")}</pre>
        </article>
      `).join("")}
    </div>
  `;
}

function renderContestStatusHint(contest) {
  const now = Date.now();
  const start = new Date(contest.start_time).getTime();
  const end = new Date(contest.end_time).getTime();
  if (Number.isNaN(start) || Number.isNaN(end)) {
    return `${escapeHTML(contest.start_time)} to ${escapeHTML(contest.end_time)}`;
  }
  if (now < start) {
    return `starts in ${formatDuration(start - now)} - ${escapeHTML(contest.start_time)}`;
  }
  if (now < end) {
    return `ends in ${formatDuration(end - now)} - ${escapeHTML(contest.end_time)}`;
  }
  return `ended - ${escapeHTML(contest.end_time)}`;
}

function formatDuration(ms) {
  const totalMinutes = Math.max(0, Math.floor(ms / 60000));
  const days = Math.floor(totalMinutes / (24 * 60));
  const hours = Math.floor((totalMinutes % (24 * 60)) / 60);
  const minutes = totalMinutes % 60;
  if (days > 0) return `${days}d ${hours}h ${minutes}m`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

function startContestPolling(contestID, contestStatus) {
  stopContestPolling();
  if (contestStatus !== "running") {
    return;
  }
  state.contestPollTimer = window.setInterval(() => {
    const hash = getCurrentHashPath().split("?")[0];
    if (hash !== `/contests/${contestID}`) {
      stopContestPolling();
      return;
    }
    renderContestDetail(contestID);
  }, 15000);
}

function stopContestPolling() {
  if (state.contestPollTimer) {
    window.clearInterval(state.contestPollTimer);
    state.contestPollTimer = null;
  }
}



function renderContestDeck(contests) {
  if (!contests || !contests.length) {
    return `<p class="view-subtitle">No contest scheduled yet.</p>`;
  }
  return `
    <div class="detail-grid">
      ${contests.map((item) => `
        <article class="detail-card">
          <div class="view-header">
            <div>
              <h3>${escapeHTML(item.title)}</h3>
              <p class="view-subtitle">${escapeHTML(item.rule_type)} - ${renderContestStatusHint(item)}</p>
            </div>
            <span class="status-pill ${contestStatusClass(item.status)}">${escapeHTML(item.status)}</span>
          </div>
          <div class="metric-list">
            <div class="metric"><span class="metric-label">Problems</span><span class="metric-value">${item.problem_count}</span></div>
            <div class="metric"><span class="metric-label">Registrations</span><span class="metric-value">${item.registered_count}</span></div>
          </div>
          <div style="margin-top:14px;">
            <a class="ghost-button" href="#/contests/${item.id}">Open Contest</a>
          </div>
        </article>
      `).join("")}
    </div>
  `;
}

function renderContestStatusBreakdown(contests) {
  const counts = { upcoming: 0, running: 0, ended: 0 };
  (contests || []).forEach((item) => {
    if (counts[item.status] !== undefined) counts[item.status] += 1;
  });
  return `
    <div class="verdict-summary-grid">
      <div class="verdict-summary-card"><span class="status-pill status-neutral">Upcoming</span><strong>${counts.upcoming}</strong></div>
      <div class="verdict-summary-card"><span class="status-pill status-pending">Running</span><strong>${counts.running}</strong></div>
      <div class="verdict-summary-card"><span class="status-pill status-accepted">Ended</span><strong>${counts.ended}</strong></div>
    </div>
  `;
}
