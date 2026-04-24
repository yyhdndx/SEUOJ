// Submission domain pages and helpers
async function renderMySubmissions() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login to view your submissions.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading submissions...</p></div>`;
  await ensureProblemTitleMap();
  await loadAndRenderMySubmissions();
  startSubmissionsPolling();
}

async function loadAndRenderMySubmissions() {
  const queryString = getCurrentHashPath().split("?")[1] || "";
  const params = new URLSearchParams(queryString);
  const page = params.get("page") || "1";
  const pageSize = params.get("page_size") || "50";
  const problemID = params.get("problem_id") || "";
  const contestID = params.get("contest_id") || "";
  const status = params.get("status") || "";

  const query = new URLSearchParams({
    page,
    page_size: pageSize,
  });
  if (problemID) query.set("problem_id", problemID);
  if (contestID) query.set("contest_id", contestID);
  if (status) query.set("status", status);

  const data = await apiFetch(`/submissions/my?${query.toString()}`, { method: "GET" });
  state.submissions = data.list || [];

  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">My Submissions</h1>
        <p class="view-subtitle">Recent results refresh automatically while judging is active. Contest-tagged runs can be filtered here too.</p>
      </div>
      <div class="mono">${hasActiveSubmission() ? "auto refresh: 3s" : "auto refresh: off"}</div>
    </div>
    <form id="submission-filters" class="toolbar">
      <div>
        <label class="field-label">Problem ID</label>
        <input class="text-input" name="problem_id" value="${escapeHTML(problemID)}" placeholder="e.g. 2" />
      </div>
      <div>
        <label class="field-label">Contest ID</label>
        <input class="text-input" name="contest_id" value="${escapeHTML(contestID)}" placeholder="optional" />
      </div>
      <div>
        <label class="field-label">Status</label>
        <select class="select-input" name="status">
          <option value="">All</option>
          ${["Pending", "Running", "Accepted", "Wrong Answer", "Compile Error", "Runtime Error", "Time Limit Exceeded", "System Error"]
            .map((item) => `<option value="${escapeHTML(item)}" ${item === status ? "selected" : ""}>${escapeHTML(item)}</option>`)
            .join("")}
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
        <button class="ghost-button" type="button" id="clear-submission-filters">Clear</button>
      </div>
    </form>
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Problem</th>
          <th>Title</th>
          <th>Language</th>
          <th>Contest</th>
          <th>Status</th>
          <th>Passed</th>
          <th>Runtime</th>
          <th>Created</th>
        </tr>
      </thead>
        <tbody>
          ${state.submissions.map((item) => `
            <tr>
              <td><a class="table-link" href="#/submissions/${item.id}">${item.id}</a></td>
              <td>${item.problem_id}</td>
              <td>${escapeHTML(state.problemTitleMap[item.problem_id] || "-")}</td>
              <td>${escapeHTML(item.language)}</td>
              <td>${item.contest_id ? `<div style="display:flex; gap:6px; align-items:center; flex-wrap:wrap;"><a class="table-link" href="#/contests/${item.contest_id}">#${item.contest_id}</a>${renderContestModeBadge(item)}</div>` : "-"}</td>
              <td><span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span></td>
              <td>${item.passed_count}/${item.total_count}</td>
            <td>${item.runtime_ms ?? "-"}</td>
            <td class="mono">${escapeHTML(item.created_at)}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;

  document.getElementById("submission-filters").addEventListener("submit", (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const next = new URLSearchParams();
    const nextProblemID = (form.get("problem_id") || "").toString().trim();
    const nextContestID = (form.get("contest_id") || "").toString().trim();
    const nextStatus = (form.get("status") || "").toString().trim();
    const nextPage = (form.get("page") || "1").toString().trim() || "1";
    const nextPageSize = (form.get("page_size") || "50").toString().trim() || "50";

    next.set("page", nextPage);
    next.set("page_size", nextPageSize);
    if (nextProblemID) next.set("problem_id", nextProblemID);
    if (nextContestID) next.set("contest_id", nextContestID);
    if (nextStatus) next.set("status", nextStatus);

    location.hash = `#/submissions?${next.toString()}`;
  });

  document.getElementById("clear-submission-filters").addEventListener("click", () => {
    location.hash = "#/submissions";
  });
}

function startSubmissionsPolling() {
  stopSubmissionsPolling();
  if (!hasActiveSubmission()) {
    return;
  }

  state.submissionsPollTimer = window.setInterval(async () => {
    if (getCurrentHashPath().split("?")[0] !== "/submissions") {
      stopSubmissionsPolling();
      return;
    }

    try {
      await loadAndRenderMySubmissions();
      if (!hasActiveSubmission()) {
        stopSubmissionsPolling();
      }
    } catch (err) {
      stopSubmissionsPolling();
      setFlash(`Auto refresh failed: ${err.message}`, true);
    }
  }, 3000);
}

function stopSubmissionsPolling() {
  if (state.submissionsPollTimer) {
    window.clearInterval(state.submissionsPollTimer);
    state.submissionsPollTimer = null;
  }
}

function hasActiveSubmission() {
  return state.submissions.some((item) => isSubmissionPollingStatus(item.status));
}

async function ensureProblemTitleMap() {
  if (Object.keys(state.problemTitleMap).length > 0) {
    return;
  }

  try {
    const data = await apiFetch("/problems?page=1&page_size=100", { method: "GET" });
    const list = data.list || [];
    state.problemTitleMap = Object.fromEntries(list.map((item) => [item.id, item.title]));
  } catch {
    state.problemTitleMap = {};
  }
}


async function renderSubmissionDetail(id) {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login to view submission detail.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading submission...</p></div>`;
  try {
    const detail = await fetchSubmissionDetail(id);
    renderSubmissionDetailView(detail);
    syncSubmissionPolling(id, detail.status);
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load submission failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function fetchSubmissionDetail(id) {
  const detail = await apiFetch(`/submissions/${id}`, { method: "GET" });
  state.submissionDetail = detail;
  return detail;
}

function renderSubmissionDetailView(detail) {
  const verdictTone = getVerdictTone(detail.status);
  const resultSummary = summarizeSubmissionResults(detail.results || []);
  const firstFailedTestcaseID = getFirstFailedTestcaseID(detail.results || []);
  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">Submission #${detail.id}</h1>
        <p class="view-subtitle">${detail.contest_id ? `Contest ${detail.contest_id} / Problem ${detail.problem_id}` : `Problem ${detail.problem_id}`} / ${escapeHTML(detail.language)}${detail.contest_id ? ` / ${detail.is_practice ? "practice" : "official"}` : ""}</p>
      </div>
      <div style="display:flex; gap:10px;">
        <button class="ghost-button" id="reuse-code-btn">Reuse Code</button>
        ${detail.contest_id ? `<a class="ghost-button" href="#/contests/${detail.contest_id}">Back to Contest</a>` : ""}
        <a class="ghost-button" href="${detail.contest_id ? `#/contests/${detail.contest_id}/problems/${detail.problem_id}` : `#/problems/${detail.problem_id}`}">${detail.contest_id ? "Back to Contest Problem" : "Back to Problem"}</a>
        <a class="ghost-button" href="#/submissions${detail.contest_id ? `?contest_id=${detail.contest_id}` : ``}">Back</a>
      </div>
    </div>
    <section class="verdict-banner ${verdictTone}">
      <div>
        <h2 class="verdict-title">${escapeHTML(detail.status)}</h2>
        <p class="verdict-subtitle">${detail.contest_id ? `${detail.is_practice ? "Practice" : "Official"} contest submission for problem ${detail.problem_id}` : `Submission for problem ${detail.problem_id}`}, created at ${escapeHTML(detail.created_at)}</p>
      </div>
      <div class="verdict-stats">
        <div class="verdict-stat">
          <span class="verdict-stat-label">Passed</span>
          <span class="verdict-stat-value">${detail.passed_count}/${detail.total_count}</span>
        </div>
        <div class="verdict-stat">
          <span class="verdict-stat-label">Runtime</span>
          <span class="verdict-stat-value">${detail.runtime_ms ?? "-" } ms</span>
        </div>
        <div class="verdict-stat">
          <span class="verdict-stat-label">Memory</span>
          <span class="verdict-stat-value">${detail.memory_kb ?? "-" } KB</span>
        </div>
        <div class="verdict-stat">
          <span class="verdict-stat-label">Judged</span>
          <span class="verdict-stat-value mono">${escapeHTML(detail.judged_at || "-")}</span>
        </div>
      </div>
    </section>
    ${Array.isArray(detail.results) && detail.results.length ? `
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Result Summary</h3>
            <p class="view-subtitle">Quick aggregate of testcase verdict distribution.</p>
          </div>
        </div>
        <div class="verdict-summary-grid">
          ${renderSummaryCard("Accepted", resultSummary.accepted, "status-accepted")}
          ${renderSummaryCard("Wrong Answer", resultSummary.wrongAnswer, "status-wrong")}
          ${renderSummaryCard("Runtime / TLE", resultSummary.runtimeLike, "status-error")}
          ${renderSummaryCard("Other", resultSummary.other, "status-pending")}
        </div>
      </section>
    ` : ""}
    <section class="detail-grid">
      <article class="detail-card">
        ${renderProblemBlock("Source Code", detail.code || "")}
        ${detail.compile_info ? renderNoticeBlock("Compile Info", detail.compile_info) : ""}
        ${detail.error_message ? renderNoticeBlock("Error Message", detail.error_message) : ""}
        ${Array.isArray(detail.results) && detail.results.length ? `
          <div class="detail-block">
            <h3>Judge Results</h3>
            <table class="data-table">
              <thead>
                <tr>
                  <th>Case</th>
                  <th>Status</th>
                  <th>Runtime</th>
                  <th>Error</th>
                </tr>
              </thead>
              <tbody>
                ${detail.results.map((item) => `
                  <tr class="${submissionResultRowClass(item, firstFailedTestcaseID)}">
                    <td>
                      ${item.testcase_id}
                      ${item.testcase_id === firstFailedTestcaseID ? '<span class="fail-badge">first fail</span>' : ""}
                    </td>
                    <td><span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                    <td>${item.runtime_ms ?? "-"}</td>
                    <td>${escapeHTML(item.error_message || "-")}</td>
                  </tr>
                `).join("")}
              </tbody>
            </table>
          </div>
        ` : ""}
      </article>
      <aside class="detail-card">
        <h3>Verdict</h3>
        <p class="view-subtitle">${isSubmissionPollingStatus(detail.status) ? "Auto polling every 2s" : "Polling stopped"}</p>
        <div class="metric-list">
          ${detail.contest_id ? `<div class="metric"><span class="metric-label">Contest</span><span class="metric-value"><a class="table-link" href="#/contests/${detail.contest_id}">#${detail.contest_id}</a></span></div>` : ""}
          ${detail.contest_id ? `<div class="metric"><span class="metric-label">Mode</span><span class="metric-value">${detail.is_practice ? "Practice" : "Official"}</span></div>` : ""}
          <div class="metric"><span class="metric-label">Status</span><span class="metric-value"><span class="status-pill ${statusClass(detail.status)}">${escapeHTML(detail.status)}</span></span></div>
          <div class="metric"><span class="metric-label">Passed Cases</span><span class="metric-value">${detail.passed_count}/${detail.total_count}</span></div>
          <div class="metric"><span class="metric-label">Runtime</span><span class="metric-value">${detail.runtime_ms ?? "-"} ms</span></div>
          <div class="metric"><span class="metric-label">Memory</span><span class="metric-value">${detail.memory_kb ?? "-"} KB</span></div>
          <div class="metric"><span class="metric-label">Created</span><span class="metric-value mono">${escapeHTML(detail.created_at)}</span></div>
          <div class="metric"><span class="metric-label">Judged</span><span class="metric-value mono">${escapeHTML(detail.judged_at || "-")}</span></div>
        </div>
      </aside>
    </section>
  `;

  document.getElementById("reuse-code-btn").addEventListener("click", () => {
    saveSubmissionDraft(detail.problem_id, detail.language, detail.code || "");
    setFlash(`Draft saved from submission #${detail.id}`, false);
    location.hash = detail.contest_id ? `#/contests/${detail.contest_id}/problems/${detail.problem_id}` : `#/problems/${detail.problem_id}`;
  });
}

function renderSummaryCard(label, value, pillClass) {
  return `
    <div class="verdict-summary-card">
      <span class="status-pill ${pillClass}">${escapeHTML(label)}</span>
      <strong>${value}</strong>
    </div>
  `;
}

function renderHomeActionForLatestSubmission(submission) {
  const label = isSubmissionPollingStatus(submission.status) ? "Track Verdict" : "Reuse and Resubmit";
  return `<button class="ghost-button" id="latest-submission-action" type="button">${label}</button>`;
}

function handleContinueLatestSubmission(submission) {
  if (isSubmissionPollingStatus(submission.status)) {
    location.hash = `#/submissions/${submission.id}`;
    return;
  }
  location.hash = `#/submissions/${submission.id}`;
  setTimeout(() => {
    const btn = document.getElementById("reuse-code-btn");
    if (btn) {
      btn.click();
    }
  }, 120);
}

function renderLatestFailureSummary(submission, submissionDetail) {
  if (!submission || isSubmissionPollingStatus(submission.status) || submission.status === "Accepted") {
    return "";
  }

  const primaryReason =
    submissionDetail?.error_message ||
    submissionDetail?.compile_info ||
    firstFailedResultMessage(submissionDetail?.results || []) ||
    "No detailed reason recorded.";

  return `
    <div class="notice-card" style="margin-top:14px;">
      <strong style="display:block; margin-bottom:8px;">Latest Failure Reason</strong>
      <div class="mono">${escapeHTML(primaryReason)}</div>
    </div>
  `;
}

function firstFailedResultMessage(results) {
  const failed = results.find((item) => item.status !== "Accepted");
  if (!failed) {
    return "";
  }
  const detail = failed.error_message ? `: ${failed.error_message}` : "";
  return `Testcase ${failed.testcase_id} ${failed.status}${detail}`;
}

function renderNoticeBlock(title, content) {
  return `
    <div class="detail-block">
      <h3>${escapeHTML(title)}</h3>
      <div class="notice-card">
        <pre>${escapeHTML(content || "")}</pre>
      </div>
    </div>
  `;
}

function summarizeSubmissionResults(results) {
  const summary = {
    accepted: 0,
    wrongAnswer: 0,
    runtimeLike: 0,
    other: 0,
  };

  results.forEach((item) => {
    if (item.status === "Accepted") {
      summary.accepted += 1;
      return;
    }
    if (item.status === "Wrong Answer") {
      summary.wrongAnswer += 1;
      return;
    }
    if (item.status === "Runtime Error" || item.status === "Time Limit Exceeded") {
      summary.runtimeLike += 1;
      return;
    }
    summary.other += 1;
  });

  return summary;
}

function getFirstFailedTestcaseID(results) {
  const failed = results.find((item) => item.status !== "Accepted");
  return failed ? failed.testcase_id : null;
}

function submissionResultRowClass(item, firstFailedTestcaseID) {
  if (item.status === "Accepted") {
    return "row-accepted";
  }
  if (item.testcase_id === firstFailedTestcaseID) {
    return "row-first-failed";
  }
  return "row-failed";
}

function getVerdictTone(status) {
  if (status === "Accepted") return "accepted";
  if (status === "Pending" || status === "Running") return "pending";
  return "error";
}

function syncSubmissionPolling(id, status) {
  stopSubmissionPolling();
  if (!isSubmissionPollingStatus(status)) {
    return;
  }

  state.submissionPollTimer = window.setInterval(async () => {
    if (!getCurrentHashPath().startsWith(`/submissions/${id}`)) {
      stopSubmissionPolling();
      return;
    }

    try {
      const latest = await fetchSubmissionDetail(id);
      renderSubmissionDetailView(latest);
      if (!isSubmissionPollingStatus(latest.status)) {
        stopSubmissionPolling();
        setFlash(`Submission #${id} finished: ${latest.status}`, false);
      }
    } catch (err) {
      stopSubmissionPolling();
      setFlash(`Auto polling failed: ${err.message}`, true);
    }
  }, 2000);
}

function stopSubmissionPolling() {
  if (state.submissionPollTimer) {
    window.clearInterval(state.submissionPollTimer);
    state.submissionPollTimer = null;
  }
}

function isSubmissionPollingStatus(status) {
  return status === "Pending" || status === "Running";
}

function statusClass(status) {
  if (status === "Accepted") return "status-accepted";
  if (status === "Pending" || status === "Running") return "status-pending";
  if (status === "Wrong Answer") return "status-wrong";
  return "status-error";
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function submissionDraftKey(problemID) {
  return `seuoj_submission_draft_${problemID}`;
}

function saveSubmissionDraft(problemID, language, code) {
  localStorage.setItem(submissionDraftKey(problemID), JSON.stringify({
    language,
    code,
    saved_at: new Date().toISOString(),
  }));
}

function readSubmissionDraft(problemID) {
  try {
    const raw = localStorage.getItem(submissionDraftKey(problemID));
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function clearSubmissionDraft(problemID) {
  localStorage.removeItem(submissionDraftKey(problemID));
}



