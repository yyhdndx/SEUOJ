const state = {
  apiBase: localStorage.getItem("seuoj_api_base") || `${location.origin}/api`,
  token: localStorage.getItem("seuoj_token") || "",
  user: readJSON("seuoj_user"),
  problems: [],
  problemDetail: null,
  submissions: [],
  submissionDetail: null,
  runResult: null,
  submissionPollTimer: null,
  submissionsPollTimer: null,
  problemTitleMap: {},
  workbenchLeftWidth: Number(localStorage.getItem("seuoj_workbench_left_width")) || 48,
  runResultHeight: Number(localStorage.getItem("seuoj_run_result_height")) || 180,
};

const submissionLanguageOptions = [
  ["cpp", "C++17"],
  ["c", "C"],
  ["python3", "Python 3"],
  ["java", "Java"],
  ["go", "Go"],
  ["rust", "Rust"],
];

const submissionLanguageTemplates = {
  cpp: "#include <iostream>\nusing namespace std;\n\nint main() {\n    int a, b;\n    cin >> a >> b;\n    cout << a + b << '\\n';\n    return 0;\n}\n",
  c: "#include <stdio.h>\n\nint main(void) {\n    int a, b;\n    scanf(\"%d %d\", &a, &b);\n    printf(\"%d\\n\", a + b);\n    return 0;\n}\n",
  python3: "a, b = map(int, input().split())\nprint(a + b)\n",
  java: "import java.util.*;\n\npublic class Main {\n    public static void main(String[] args) {\n        Scanner in = new Scanner(System.in);\n        int a = in.nextInt();\n        int b = in.nextInt();\n        System.out.println(a + b);\n    }\n}\n",
  go: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    var a, b int\n    fmt.Scan(&a, &b)\n    fmt.Println(a + b)\n}\n",
  rust: "use std::io::{self, Read};\n\nfn main() {\n    let mut input = String::new();\n    io::stdin().read_to_string(&mut input).unwrap();\n    let nums: Vec<i32> = input.split_whitespace().map(|s| s.parse().unwrap()).collect();\n    println!(\"{}\", nums[0] + nums[1]);\n}\n",
};

const app = document.getElementById("app");
const flash = document.getElementById("flash");
const accountSlot = document.getElementById("account-slot");

window.addEventListener("hashchange", renderRoute);
window.addEventListener("beforeunload", () => {
  stopSubmissionPolling();
  stopSubmissionsPolling();
});

document.addEventListener("DOMContentLoaded", async () => {
  renderAccount();
  updateNavActive();
  if (state.token && !state.user) {
    await refreshCurrentUser();
  }
  if (!location.hash) {
    location.hash = "#/problems";
    return;
  }
  renderRoute();
});

function readJSON(key) {
  const raw = localStorage.getItem(key);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function persistAuth(token, user) {
  state.token = token || "";
  state.user = user || null;

  if (state.token) {
    localStorage.setItem("seuoj_token", state.token);
  } else {
    localStorage.removeItem("seuoj_token");
  }

  if (state.user) {
    localStorage.setItem("seuoj_user", JSON.stringify(state.user));
  } else {
    localStorage.removeItem("seuoj_user");
  }

  renderAccount();
}

async function apiFetch(path, options = {}) {
  const headers = { ...(options.headers || {}) };
  if (!(options.body instanceof FormData)) {
    headers["Content-Type"] = headers["Content-Type"] || "application/json";
  }
  if (state.token) {
    headers.Authorization = `Bearer ${state.token}`;
  }

  const response = await fetch(`${state.apiBase}${path}`, {
    ...options,
    headers,
  });

  const data = await response.json().catch(() => ({
    code: 1,
    message: `invalid response: ${response.status}`,
    data: null,
  }));

  if (data.code !== 0) {
    throw new Error(data.message || "request failed");
  }
  return data.data;
}

async function refreshCurrentUser() {
  try {
    const user = await apiFetch("/auth/me", { method: "GET" });
    persistAuth(state.token, user);
  } catch (err) {
    persistAuth("", null);
    setFlash(`Login expired: ${err.message}`, true);
  }
}

function setFlash(message, isError = false) {
  if (!message) {
    flash.className = "flash hidden";
    flash.textContent = "";
    return;
  }
  flash.className = `flash ${isError ? "error" : ""}`;
  flash.textContent = message;
}

function renderAccount() {
  const adminLink = document.getElementById("admin-link");
  const adminManageLink = document.getElementById("admin-manage-link");
  if (adminLink) {
    adminLink.style.display = state.user?.role === "admin" ? "inline-block" : "none";
  }
  if (adminManageLink) {
    adminManageLink.style.display = state.user?.role === "admin" ? "inline-block" : "none";
  }

  if (!state.user) {
    accountSlot.innerHTML = `<span class="mono">guest</span><a class="link-button" href="#/auth">login</a>`;
    return;
  }

  accountSlot.innerHTML = `
    <span class="mono">${escapeHTML(state.user.username)}</span>
    <span class="status-pill ${state.user.role === "admin" ? "status-accepted" : "status-pending"}">${escapeHTML(state.user.role)}</span>
    <button class="ghost-button" id="logout-btn">Logout</button>
  `;

  document.getElementById("logout-btn").addEventListener("click", () => {
    persistAuth("", null);
    setFlash("Logged out", false);
    location.hash = "#/auth";
  });
}

function getCurrentHashPath() {
  return location.hash.replace(/^#/, "") || "/";
}

async function renderRoute() {
  stopSubmissionPolling();
  stopSubmissionsPolling();
  setFlash("");
  updateNavActive();

  const hash = getCurrentHashPath();
  document.body.classList.toggle("problem-fullscreen", hash.startsWith("/problems/"));

  if (hash === "/") {
    return renderHome();
  }
  if (hash.startsWith("/problems/")) {
    return renderProblemDetail(hash.split("/")[2]);
  }
  if (hash.startsWith("/problems")) {
    return renderProblems();
  }
  if (hash === "/submissions") {
    return renderMySubmissions();
  }
  if (hash.startsWith("/submissions/")) {
    return renderSubmissionDetail(hash.split("/")[2]);
  }
  if (hash === "/admin/problems") {
    return renderAdminProblems();
  }
  if (hash === "/admin/problems/new") {
    return renderAdminProblemCreate();
  }
  if (hash.startsWith("/admin/problems/") && hash.endsWith("/edit")) {
    return renderAdminProblemEdit(hash.split("/")[3]);
  }
  if (hash.startsWith("/admin/problems/") && !hash.endsWith("/edit")) {
    return renderAdminProblemDetail(hash.split("/")[3]);
  }
  if (hash === "/auth") {
    return renderAuth();
  }

  app.innerHTML = `<div class="detail-card"><h2>Not Found</h2><p>Unknown route: ${escapeHTML(hash)}</p></div>`;
}

function updateNavActive() {
  const hash = getCurrentHashPath();
  const links = document.querySelectorAll(".topnav a");
  links.forEach((link) => {
    const href = link.getAttribute("href") || "";
    const route = href.replace(/^#/, "") || "/";
    const active =
      route === "/"
        ? hash === "/"
        : hash === route || hash.startsWith(`${route}/`) || hash.startsWith(`${route}?`);
    link.classList.toggle("active", active);
  });
}

function renderAuth() {
  app.innerHTML = `
    <section class="split-hero">
      <div class="hero-card">
        <h1 class="view-title">Account Portal</h1>
        <p class="view-subtitle">Minimal auth loop for problem solving and judging.</p>
      </div>
      <div class="hero-card">
        <h3>Current State</h3>
        <p class="mono">${state.user ? `${escapeHTML(state.user.username)} / ${escapeHTML(state.user.role)}` : "guest"}</p>
      </div>
    </section>
    <div class="grid-form" style="margin-top:18px;">
      <form id="login-form" class="detail-card">
        <h3>Login</h3>
        <label class="field-label">Username</label>
        <input class="text-input" name="username" required />
        <label class="field-label">Password</label>
        <input class="text-input" name="password" type="password" required />
        <div style="margin-top:14px;">
          <button class="primary-button" type="submit">Login</button>
        </div>
      </form>
      <form id="register-form" class="detail-card">
        <h3>Register</h3>
        <label class="field-label">Username</label>
        <input class="text-input" name="username" required />
        <label class="field-label">Student ID</label>
        <input class="text-input" name="userid" required />
        <label class="field-label">Password</label>
        <input class="text-input" name="password" type="password" required />
        <div style="margin-top:14px;">
          <button class="primary-button" type="submit">Register</button>
        </div>
      </form>
    </div>
  `;

  document.getElementById("login-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      const data = await apiFetch("/auth/login", {
        method: "POST",
        body: JSON.stringify({
          username: form.get("username"),
          password: form.get("password"),
        }),
      });
      persistAuth(data.token, data.user);
      setFlash("Login success", false);
      location.hash = "#/problems";
    } catch (err) {
      setFlash(err.message, true);
    }
  });

  document.getElementById("register-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await apiFetch("/auth/register", {
        method: "POST",
        body: JSON.stringify({
          username: form.get("username"),
          userid: form.get("userid"),
          password: form.get("password"),
        }),
      });
      setFlash("Register success, now login", false);
    } catch (err) {
      setFlash(err.message, true);
    }
  });
}

async function renderHome() {
  app.innerHTML = `<div class="detail-card"><p>Loading dashboard...</p></div>`;

  let latestProblems = [];
  let latestSubmissions = [];
  let latestSubmission = null;
  let latestSubmissionDetail = null;
  let latestAcceptedSubmission = null;
  let totalProblems = 0;

  try {
    const problemsData = await apiFetch("/problems?page=1&page_size=5", { method: "GET" });
    latestProblems = problemsData.list || [];
    totalProblems = problemsData.total || latestProblems.length;
  } catch {
    latestProblems = [];
    totalProblems = 0;
  }

  if (state.token) {
    try {
      const submissionsData = await apiFetch("/submissions/my?page=1&page_size=5", { method: "GET" });
      latestSubmissions = submissionsData.list || [];
      latestSubmission = latestSubmissions[0] || null;
      latestAcceptedSubmission = latestSubmissions.find((item) => item.status === "Accepted") || null;
      if (latestSubmission) {
        latestSubmissionDetail = await apiFetch(`/submissions/${latestSubmission.id}`, { method: "GET" });
      }
    } catch {
      latestSubmissions = [];
      latestSubmission = null;
      latestSubmissionDetail = null;
      latestAcceptedSubmission = null;
    }
  }

  app.innerHTML = `
    <section class="split-hero">
      <div class="hero-card">
        <h1 class="view-title">SEU OJ</h1>
        <p class="view-subtitle">A compact online judge focused on the shortest path from reading a problem to getting a verdict.</p>
        <div style="display:flex; gap:10px; margin-top:18px; flex-wrap:wrap;">
          <a class="primary-button" href="#/problems">Browse Problems</a>
          <a class="ghost-button" href="#/submissions">My Submissions</a>
          ${latestSubmission ? `<button class="ghost-button" id="continue-last-work-btn">Continue Last Work</button>` : ""}
          ${state.user?.role === "admin" ? `<a class="ghost-button" href="#/admin/problems">Manage Problems</a>` : ""}
        </div>
      </div>
      <div class="hero-card">
        <h3>Account Snapshot</h3>
        <div class="metric-list">
          <div class="metric"><span class="metric-label">User</span><span class="metric-value">${escapeHTML(state.user?.username || "guest")}</span></div>
          <div class="metric"><span class="metric-label">Role</span><span class="metric-value">${escapeHTML(state.user?.role || "guest")}</span></div>
          <div class="metric"><span class="metric-label">Problems</span><span class="metric-value">${totalProblems}</span></div>
          <div class="metric"><span class="metric-label">Recent Submissions</span><span class="metric-value">${latestSubmissions.length}</span></div>
          <div class="metric"><span class="metric-label">Token</span><span class="metric-value mono">${state.token ? "loaded" : "missing"}</span></div>
          <div class="metric"><span class="metric-label">API Base</span><span class="metric-value mono">${escapeHTML(state.apiBase)}</span></div>
        </div>
      </div>
    </section>
    <section class="detail-grid" style="margin-top:18px;">
      <article class="detail-card">
        <div class="view-header">
          <div>
            <h3>Latest Problems</h3>
            <p class="view-subtitle">Recently published visible problems.</p>
          </div>
          <a class="ghost-button" href="#/problems">All Problems</a>
        </div>
        ${latestProblems.length ? `
          <table class="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Title</th>
                <th>Time</th>
                <th>Memory</th>
              </tr>
            </thead>
            <tbody>
              ${latestProblems.map((item) => `
                <tr>
                  <td>${item.id}</td>
                  <td><a class="table-link" href="#/problems/${item.id}">${escapeHTML(item.title)}</a></td>
                  <td>${item.time_limit_ms} ms</td>
                  <td>${item.memory_limit_mb} MB</td>
                </tr>
              `).join("")}
            </tbody>
          </table>
        ` : `<p>No problem data available.</p>`}
      </article>
      <aside class="detail-card">
        <div class="view-header">
          <div>
            <h3>Recent Submissions</h3>
            <p class="view-subtitle">Visible after login.</p>
          </div>
          <a class="ghost-button" href="#/submissions">All Submissions</a>
        </div>
        ${state.token ? (
          latestSubmissions.length ? `
            <table class="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Status</th>
                  <th>Passed</th>
                </tr>
              </thead>
              <tbody>
                ${latestSubmissions.map((item) => `
                  <tr>
                    <td><a class="table-link" href="#/submissions/${item.id}">${item.id}</a></td>
                    <td><span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                    <td>${item.passed_count}/${item.total_count}</td>
                  </tr>
                `).join("")}
              </tbody>
            </table>
          ` : `<p>No submission yet.</p>`
        ) : `<p>Please login to view your recent submissions.</p>`}
      </aside>
    </section>
    ${latestSubmission ? `
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Latest Submission Focus</h3>
            <p class="view-subtitle">A quick summary of your most recent judge result.</p>
          </div>
          <a class="ghost-button" href="#/submissions/${latestSubmission.id}">Open Detail</a>
        </div>
        <div class="verdict-banner ${getVerdictTone(latestSubmission.status)}">
          <div>
            <h2 class="verdict-title">${escapeHTML(latestSubmission.status)}</h2>
            <p class="verdict-subtitle">Submission #${latestSubmission.id} for problem ${latestSubmission.problem_id}</p>
            ${renderLatestFailureSummary(latestSubmission, latestSubmissionDetail)}
          </div>
          <div class="verdict-stats">
            <div class="verdict-stat">
              <span class="verdict-stat-label">Passed</span>
              <span class="verdict-stat-value">${latestSubmission.passed_count}/${latestSubmission.total_count}</span>
            </div>
            <div class="verdict-stat">
              <span class="verdict-stat-label">Runtime</span>
              <span class="verdict-stat-value">${latestSubmission.runtime_ms ?? "-"} ms</span>
            </div>
            <div class="verdict-stat">
              <span class="verdict-stat-label">Created</span>
              <span class="verdict-stat-value mono">${escapeHTML(latestSubmission.created_at)}</span>
            </div>
            <div class="verdict-stat">
              <span class="verdict-stat-label">Action</span>
              <span class="verdict-stat-value">
                ${renderHomeActionForLatestSubmission(latestSubmission)}
              </span>
            </div>
          </div>
        </div>
      </section>
    ` : ""}
    ${latestAcceptedSubmission ? `
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Latest Accepted Problem</h3>
            <p class="view-subtitle">Your most recent successful submission.</p>
          </div>
          <a class="ghost-button" href="#/submissions/${latestAcceptedSubmission.id}">Open Accepted Submission</a>
        </div>
        <div class="verdict-banner accepted">
          <div>
            <h2 class="verdict-title">Accepted</h2>
            <p class="verdict-subtitle">Problem ${latestAcceptedSubmission.problem_id} / Submission #${latestAcceptedSubmission.id}</p>
          </div>
          <div class="verdict-stats">
            <div class="verdict-stat">
              <span class="verdict-stat-label">Problem</span>
              <span class="verdict-stat-value">${escapeHTML(state.problemTitleMap[latestAcceptedSubmission.problem_id] || `#${latestAcceptedSubmission.problem_id}`)}</span>
            </div>
            <div class="verdict-stat">
              <span class="verdict-stat-label">Passed</span>
              <span class="verdict-stat-value">${latestAcceptedSubmission.passed_count}/${latestAcceptedSubmission.total_count}</span>
            </div>
            <div class="verdict-stat">
              <span class="verdict-stat-label">Runtime</span>
              <span class="verdict-stat-value">${latestAcceptedSubmission.runtime_ms ?? "-"} ms</span>
            </div>
            <div class="verdict-stat">
              <span class="verdict-stat-label">Continue</span>
              <span class="verdict-stat-value">
                <button class="ghost-button" id="open-accepted-problem-btn" type="button">Open Problem</button>
              </span>
            </div>
          </div>
        </div>
      </section>
    ` : ""}
    ${state.token ? `
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Personal Snapshot</h3>
            <p class="view-subtitle">A quick comparison between your latest overall result and latest accepted result.</p>
          </div>
        </div>
        <div class="verdict-summary-grid">
          <div class="verdict-summary-card">
            <span class="status-pill ${latestSubmission ? statusClass(latestSubmission.status) : "status-neutral"}">Latest</span>
            <strong>${escapeHTML(latestSubmission?.status || "N/A")}</strong>
            <div class="view-subtitle">Submission ${latestSubmission ? `#${latestSubmission.id}` : "not found"}</div>
          </div>
          <div class="verdict-summary-card">
            <span class="status-pill ${latestAcceptedSubmission ? "status-accepted" : "status-neutral"}">Latest Accepted</span>
            <strong>${latestAcceptedSubmission ? `#${latestAcceptedSubmission.id}` : "N/A"}</strong>
            <div class="view-subtitle">${latestAcceptedSubmission ? `Problem ${latestAcceptedSubmission.problem_id}` : "No accepted submission yet"}</div>
          </div>
          <div class="verdict-summary-card">
            <span class="status-pill status-neutral">Pending / Running</span>
            <strong>${latestSubmissions.filter((item) => isSubmissionPollingStatus(item.status)).length}</strong>
            <div class="view-subtitle">Within recent submissions</div>
          </div>
          <div class="verdict-summary-card">
            <span class="status-pill status-neutral">Accepted Rate</span>
            <strong>${latestSubmissions.length ? Math.round((latestSubmissions.filter((item) => item.status === "Accepted").length / latestSubmissions.length) * 100) : 0}%</strong>
            <div class="view-subtitle">Based on latest visible sample</div>
          </div>
        </div>
      </section>
    ` : ""}
    <section class="detail-card" style="margin-top:18px;">
      <div class="view-header">
        <div>
          <h3>Workflow</h3>
          <p class="view-subtitle">Recommended order for using the current system.</p>
        </div>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Step</th>
            <th>Route</th>
            <th>Purpose</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>1</td>
            <td class="mono">#/auth</td>
            <td>Register or login, load JWT into local storage.</td>
          </tr>
          <tr>
            <td>2</td>
            <td class="mono">#/problems</td>
            <td>Browse problem list and open detail pages.</td>
          </tr>
          <tr>
            <td>3</td>
            <td class="mono">#/submissions</td>
            <td>Track verdict updates and inspect details.</td>
          </tr>
          <tr>
            <td>4</td>
            <td class="mono">#/admin/problems</td>
            <td>Admin-only creation and maintenance flow.</td>
          </tr>
        </tbody>
      </table>
    </section>
  `;

  if (latestSubmission) {
    const continueBtn = document.getElementById("continue-last-work-btn");
    if (continueBtn) {
      continueBtn.addEventListener("click", () => {
        handleContinueLatestSubmission(latestSubmission);
      });
    }

    const quickAction = document.getElementById("latest-submission-action");
    if (quickAction) {
      quickAction.addEventListener("click", () => {
        handleContinueLatestSubmission(latestSubmission);
      });
    }
  }

  if (latestAcceptedSubmission) {
    const openAcceptedProblemBtn = document.getElementById("open-accepted-problem-btn");
    if (openAcceptedProblemBtn) {
      openAcceptedProblemBtn.addEventListener("click", () => {
        location.hash = `#/problems/${latestAcceptedSubmission.problem_id}`;
      });
    }
  }
}

async function renderProblems() {
  app.innerHTML = `<div class="detail-card"><p>Loading problems...</p></div>`;
  try {
    const queryString = getCurrentHashPath().split("?")[1] || "";
    const keyword = new URLSearchParams(queryString).get("keyword") || "";
    const data = await apiFetch(`/problems?page=1&page_size=50${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ""}`, {
      method: "GET",
    });
    state.problems = data.list || [];
    state.problemTitleMap = Object.fromEntries(state.problems.map((item) => [item.id, item.title]));
    const myProblemStatusMap = state.token ? await loadProblemStatusMap() : {};

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Problemset</h1>
          <p class="view-subtitle">Compact listing with fast scan and direct entry to solving.</p>
        </div>
        <form id="problem-search" class="inline-form" style="grid-template-columns: 1fr auto;">
          <input class="text-input" name="keyword" value="${escapeHTML(keyword)}" placeholder="Search by title" />
          <button class="ghost-button" type="submit">Search</button>
        </form>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Display</th>
            <th>Title</th>
            <th>My Status</th>
            <th>Mode</th>
            <th>Time</th>
            <th>Memory</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          ${state.problems.map((item) => `
            <tr>
              <td>${item.id}</td>
              <td class="mono">${escapeHTML(item.display_id || "-")}</td>
              <td><a class="table-link" href="#/problems/${item.id}">${escapeHTML(item.title)}</a></td>
              <td>${renderProblemStatusPill(myProblemStatusMap[item.id])}</td>
              <td>${escapeHTML(item.judge_mode)}</td>
              <td>${item.time_limit_ms} ms</td>
              <td>${item.memory_limit_mb} MB</td>
              <td class="mono">${escapeHTML(item.created_at)}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    `;

    document.getElementById("problem-search").addEventListener("submit", (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const nextKeyword = (form.get("keyword") || "").toString().trim();
      location.hash = nextKeyword ? `#/problems?keyword=${encodeURIComponent(nextKeyword)}` : "#/problems";
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load problems failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderProblemDetail(id) {
  app.innerHTML = `<div class="detail-card"><p>Loading problem...</p></div>`;
  try {
    const problem = await apiFetch(`/problems/${id}`, { method: "GET" });
    state.problemDetail = problem;
    if (!state.runResult || state.runResult.problemID !== problem.id) {
      state.runResult = null;
    }
    const draft = readSubmissionDraft(problem.id);
    const selectedLanguage = draft?.language || "cpp";
    const initialCode = draft?.code || getDefaultCodeTemplate(selectedLanguage);
    const sampleCases = Array.isArray(problem.testcases)
      ? problem.testcases.filter((item) => item.case_type === "sample")
      : [];
    const recentSubmissions = state.token ? await loadRecentProblemSubmissions(problem.id) : [];

    app.innerHTML = `
      <section class="problem-workbench resizable" id="problem-workbench" style="--problem-pane-width:${state.workbenchLeftWidth}%;">
        <article class="problem-pane">
          <div class="pane-header">
            <div>
              <h2 class="pane-title">${escapeHTML(problem.title)}</h2>
              <p class="view-subtitle">#${problem.id} / ${escapeHTML(problem.display_id || "-")} / ${escapeHTML(problem.judge_mode)}</p>
            </div>
            <div class="pill-row">
              <span class="status-pill status-neutral">${escapeHTML(problem.judge_mode)}</span>
            </div>
          </div>
          <div class="pane-content">
            ${renderProblemBlock("Description", problem.description)}
            ${renderProblemBlock("Input", problem.input_desc)}
            ${renderProblemBlock("Output", problem.output_desc)}
            ${renderProblemBlock("Sample Input", problem.sample_input)}
            ${renderProblemBlock("Sample Output", problem.sample_output)}
            ${sampleCases.length ? `
              <div class="detail-block">
                <h3>Sample Testcases</h3>
                <p class="view-subtitle">${sampleCases.length} sample case(s) are available.</p>
                ${renderSampleCaseList(sampleCases)}
              </div>
            ` : ""}
            ${renderProblemBlock("Source / Hint", `${problem.source || ""}\n${problem.hint || ""}`.trim())}
            ${renderProblemRecentSubmissions(problem.id, recentSubmissions)}
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
      if (!state.token) {
        setFlash("Please login before running code", true);
        location.hash = "#/auth";
        return;
      }

      const form = new FormData(document.getElementById("submit-form"));
      const code = (form.get("code") || "").toString();
      const language = (form.get("language") || "cpp").toString();
      saveSubmissionDraft(problem.id, language, code);

      try {
        const result = await apiFetch("/submissions/run", {
          method: "POST",
          body: JSON.stringify({
            problem_id: Number(problem.id),
            language,
            code,
          }),
        });
        state.runResult = { problemID: problem.id, ...result };
        renderProblemDetail(problem.id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("submit-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      if (!state.token) {
        setFlash("Please login before submitting", true);
        location.hash = "#/auth";
        return;
      }

      const form = new FormData(event.currentTarget);
      const language = (form.get("language") || "").toString();
      const code = (form.get("code") || "").toString();
      saveSubmissionDraft(problem.id, language, code);
      try {
        const result = await apiFetch("/submissions", {
          method: "POST",
          body: JSON.stringify({
            problem_id: Number(problem.id),
            language,
            code,
          }),
        });
        setFlash(`Submission created: #${result.submission_id}, status ${result.status}`, false);
        location.hash = `#/submissions/${result.submission_id}`;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load problem failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function initProblemWorkbenchUI() {
  const workbench = document.getElementById("problem-workbench");
  const divider = document.getElementById("workbench-divider");
  if (!workbench || !divider || window.innerWidth <= 960) {
    return;
  }

  let dragging = false;
  const stopDragging = () => {
    dragging = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  };

  const onMove = (event) => {
    if (!dragging) return;
    const rect = workbench.getBoundingClientRect();
    const next = ((event.clientX - rect.left) / rect.width) * 100;
    const bounded = Math.min(72, Math.max(38, next));
    state.workbenchLeftWidth = bounded;
    localStorage.setItem("seuoj_workbench_left_width", String(bounded));
    workbench.style.setProperty("--problem-pane-width", `${bounded}%`);
  };

  const onMouseUp = () => {
    stopDragging();
  };

  divider.addEventListener("mousedown", (event) => {
    event.preventDefault();
    dragging = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  });

  window.addEventListener("mousemove", onMove);
  window.addEventListener("mouseup", onMouseUp);
  window.addEventListener("mouseleave", onMouseUp);
}

function initRunResultUI() {
  const panel = document.getElementById("run-result-panel");
  const resizer = document.getElementById("run-result-resizer");
  if (!panel || !resizer) {
    return;
  }

  let dragging = false;
  let startY = 0;
  let startHeight = 0;

  const stopDragging = () => {
    dragging = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  };

  const onMove = (event) => {
    if (!dragging) return;
    const delta = event.clientY - startY;
    const nextHeight = Math.max(80, Math.min(420, startHeight - delta));
    state.runResultHeight = nextHeight;
    localStorage.setItem("seuoj_run_result_height", String(nextHeight));
    panel.style.height = `${nextHeight}px`;
  };

  resizer.addEventListener("mousedown", (event) => {
    event.preventDefault();
    dragging = true;
    startY = event.clientY;
    startHeight = panel.getBoundingClientRect().height;
    document.body.style.cursor = "ns-resize";
    document.body.style.userSelect = "none";
  });

  window.addEventListener("mousemove", onMove);
  window.addEventListener("mouseup", stopDragging);
  window.addEventListener("mouseleave", stopDragging);
}

function renderProblemBlock(title, content) {
  return `
    <div class="detail-block">
      <h3>${escapeHTML(title)}</h3>
      <pre>${escapeHTML(content || "")}</pre>
    </div>
  `;
}

function renderRunResultPanel() {
  if (!state.runResult) {
    return "";
  }

  const result = state.runResult;
  return `
    <div class="run-result-panel" id="run-result-panel" style="height:${state.runResultHeight}px;">
      <div class="run-result-resizer" id="run-result-resizer" aria-hidden="true"></div>
      <div class="run-result-head">
        <strong>Run Result</strong>
        <span class="status-pill ${statusClass(result.status)}">${escapeHTML(result.status)}</span>
      </div>
      ${result.compile_info ? `<pre class="run-result-pre">${escapeHTML(result.compile_info)}</pre>` : ""}
      ${result.error_message && !result.compile_info ? `<div class="run-result-error">${escapeHTML(result.error_message)}</div>` : ""}
      ${(result.results || []).map((item, index) => `
        <div class="run-case-card">
          <div class="run-case-head">
            <span>Sample ${index + 1}</span>
          </div>
          <div class="run-case-grid">
            <div>
              <label class="field-label">Input</label>
              <pre class="run-result-pre">${escapeHTML(item.input_data || "")}</pre>
            </div>
            <div>
              <label class="field-label">Expected</label>
              <pre class="run-result-pre">${escapeHTML(item.expected_output || "")}</pre>
            </div>
            <div class="full">
              <label class="field-label">Actual</label>
              <pre class="run-result-pre">${escapeHTML(item.actual_output || "")}</pre>
            </div>
          </div>
        </div>
      `).join("")}
    </div>
  `;
}

function getDefaultCodeTemplate(language) {
  return submissionLanguageTemplates[language] || submissionLanguageTemplates.cpp;
}

function renderSampleCaseList(sampleCases) {
  return sampleCases.map((item, index) => `
    <div class="detail-card" style="margin-top:12px; padding:14px;">
      <div class="view-header" style="margin-bottom:10px;">
        <div>
          <h3 style="margin:0;">Sample #${index + 1}</h3>
          <p class="view-subtitle">sort_order=${item.sort_order}, active=${item.is_active ? "true" : "false"}</p>
        </div>
      </div>
      <div class="grid-form">
        <div>
          <label class="field-label">Input</label>
          <pre>${escapeHTML(item.input_data || "")}</pre>
        </div>
        <div>
          <label class="field-label">Output</label>
          <pre>${escapeHTML(item.output_data || "")}</pre>
        </div>
      </div>
    </div>
  `).join("");
}

async function loadProblemStatusMap() {
  try {
    const data = await apiFetch("/submissions/my?page=1&page_size=100", { method: "GET" });
    const map = {};
    for (const item of data.list || []) {
      const current = map[item.problem_id];
      if (!current || (current !== "Accepted" && item.status === "Accepted")) {
        map[item.problem_id] = item.status;
      }
    }
    return map;
  } catch {
    return {};
  }
}

async function loadRecentProblemSubmissions(problemID) {
  try {
    const data = await apiFetch(`/submissions/my?page=1&page_size=5&problem_id=${problemID}`, { method: "GET" });
    return data.list || [];
  } catch {
    return [];
  }
}

function renderProblemStatusPill(status) {
  if (!state.token) {
    return `<span class="status-pill status-neutral">Login to track</span>`;
  }
  if (!status) {
    return `<span class="status-pill status-neutral">Not Submitted</span>`;
  }
  return `<span class="status-pill ${statusClass(status)}">${escapeHTML(status)}</span>`;
}

function renderProblemRecentSubmissions(problemID, submissions) {
  if (!state.token) {
    return `
      <div class="detail-block">
        <h3>My Recent Submissions</h3>
        <p class="view-subtitle">Login to see your attempts on this problem.</p>
      </div>
    `;
  }

  if (!submissions.length) {
    return `
      <div class="detail-block">
        <h3>My Recent Submissions</h3>
        <p class="view-subtitle">No submissions for this problem yet.</p>
      </div>
    `;
  }

  return `
    <div class="detail-block">
      <div class="view-header" style="margin-bottom:12px;">
        <div>
          <h3 style="margin:0;">My Recent Submissions</h3>
          <p class="view-subtitle">Latest attempts on this problem.</p>
        </div>
        <a class="ghost-button" href="#/submissions?problem_id=${problemID}">View All</a>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Status</th>
            <th>Passed</th>
            <th>Runtime</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          ${submissions.map((item) => `
            <tr>
              <td><a class="table-link" href="#/submissions/${item.id}">${item.id}</a></td>
              <td><span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span></td>
              <td>${item.passed_count}/${item.total_count}</td>
              <td>${item.runtime_ms ?? "-"}</td>
              <td class="mono">${escapeHTML(item.created_at)}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    </div>
  `;
}

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
  const status = params.get("status") || "";

  const query = new URLSearchParams({
    page,
    page_size: pageSize,
  });
  if (problemID) query.set("problem_id", problemID);
  if (status) query.set("status", status);

  const data = await apiFetch(`/submissions/my?${query.toString()}`, { method: "GET" });
  state.submissions = data.list || [];

  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">My Submissions</h1>
        <p class="view-subtitle">Recent results refresh automatically while judging is active.</p>
      </div>
      <div class="mono">${hasActiveSubmission() ? "auto refresh: 3s" : "auto refresh: off"}</div>
    </div>
    <form id="submission-filters" class="toolbar">
      <div>
        <label class="field-label">Problem ID</label>
        <input class="text-input" name="problem_id" value="${escapeHTML(problemID)}" placeholder="e.g. 2" />
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
    const nextStatus = (form.get("status") || "").toString().trim();
    const nextPage = (form.get("page") || "1").toString().trim() || "1";
    const nextPageSize = (form.get("page_size") || "50").toString().trim() || "50";

    next.set("page", nextPage);
    next.set("page_size", nextPageSize);
    if (nextProblemID) next.set("problem_id", nextProblemID);
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
    if (getCurrentHashPath() !== "/submissions") {
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

function renderAdminProblemCreate() {
  return renderAdminProblemForm();
}

async function renderAdminProblemEdit(id) {
  return renderAdminProblemForm(id);
}

async function renderAdminProblemDetail(id) {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before viewing admin problem detail.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access full problem detail.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading admin problem detail...</p></div>`;
  try {
    const problem = await apiFetch(`/admin/problems/${id}`, { method: "GET" });
    const samples = (problem.testcases || []).filter((item) => item.case_type === "sample");
    const hiddenCases = (problem.testcases || []).filter((item) => item.case_type === "hidden");

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(problem.title)}</h1>
          <p class="view-subtitle">Admin preview for #${problem.id} / ${escapeHTML(problem.display_id || "-")}</p>
        </div>
        <div style="display:flex; gap:10px;">
          <a class="ghost-button" href="#/admin/problems">Back to Manage</a>
          <a class="primary-button" href="#/admin/problems/${problem.id}/edit">Edit</a>
        </div>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          ${renderProblemBlock("Description", problem.description)}
          ${renderProblemBlock("Input", problem.input_desc)}
          ${renderProblemBlock("Output", problem.output_desc)}
          ${renderProblemBlock("Sample Input", problem.sample_input)}
          ${renderProblemBlock("Sample Output", problem.sample_output)}
          ${renderProblemBlock("Hint", problem.hint)}
        </article>
        <aside class="detail-card">
          <h3>Meta</h3>
          <div class="metric-list">
            <div class="metric"><span class="metric-label">Visible</span><span class="metric-value"><span class="status-pill ${problem.visible ? "status-accepted" : "status-hidden"}">${problem.visible ? "Visible" : "Hidden"}</span></span></div>
            <div class="metric"><span class="metric-label">Created By</span><span class="metric-value">${problem.created_by}</span></div>
            <div class="metric"><span class="metric-label">Judge Mode</span><span class="metric-value"><span class="status-pill status-neutral">${escapeHTML(problem.judge_mode)}</span></span></div>
            <div class="metric"><span class="metric-label">Time Limit</span><span class="metric-value">${problem.time_limit_ms} ms</span></div>
            <div class="metric"><span class="metric-label">Memory Limit</span><span class="metric-value">${problem.memory_limit_mb} MB</span></div>
            <div class="metric"><span class="metric-label">Created At</span><span class="metric-value mono">${escapeHTML(problem.created_at)}</span></div>
            <div class="metric"><span class="metric-label">Updated At</span><span class="metric-value mono">${escapeHTML(problem.updated_at)}</span></div>
          </div>
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Sample Testcases</h3>
            <p class="view-subtitle">Visible to users.</p>
          </div>
          <div class="mono">${samples.length} case(s)</div>
        </div>
        ${samples.length ? renderAdminCaseTable(samples) : "<p>No sample testcase.</p>"}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>Hidden Testcases</h3>
            <p class="view-subtitle">Judge-only data, admin visible.</p>
          </div>
          <div class="mono">${hiddenCases.length} case(s)</div>
        </div>
        ${hiddenCases.length ? renderAdminCaseTable(hiddenCases) : "<p>No hidden testcase.</p>"}
      </section>
    `;
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load admin problem detail failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAdminProblemForm(problemID = null) {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before creating problems.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access problem creation.</p></div>`;
    return;
  }

  let initialProblem = null;
  if (problemID) {
    app.innerHTML = `<div class="detail-card"><p>Loading problem editor...</p></div>`;
    try {
      initialProblem = await apiFetch(`/admin/problems/${problemID}`, { method: "GET" });
    } catch (err) {
      app.innerHTML = `<div class="detail-card"><p>Load admin problem failed: ${escapeHTML(err.message)}</p></div>`;
      return;
    }
  }

  const sampleCasesInitial = initialProblem?.testcases?.filter((item) => item.case_type === "sample") || [];
  const hiddenCasesInitial = initialProblem?.testcases?.filter((item) => item.case_type === "hidden") || [];

  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">${problemID ? "Edit Problem" : "Create Problem"}</h1>
        <p class="view-subtitle">${problemID ? "Update statement and testcase set." : "Minimal admin page for adding one problem and core testcases."}</p>
      </div>
      <div>
        <a class="ghost-button" href="#/admin/problems">Back to Manage</a>
      </div>
    </div>
    <form id="admin-problem-form" class="grid-form">
      <div>
        <label class="field-label">Display ID</label>
        <input class="text-input" name="display_id" value="${escapeHTML(initialProblem?.display_id || "1000")}" required />
      </div>
      <div>
        <label class="field-label">Title</label>
        <input class="text-input" name="title" value="${escapeHTML(initialProblem?.title || "")}" required />
      </div>
      <div class="full">
        <label class="field-label">Description</label>
        <textarea class="text-area" name="description" required>${escapeHTML(initialProblem?.description || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Input Description</label>
        <textarea class="text-area" name="input_desc">${escapeHTML(initialProblem?.input_desc || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Output Description</label>
        <textarea class="text-area" name="output_desc">${escapeHTML(initialProblem?.output_desc || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Sample Input</label>
        <textarea class="text-area" name="sample_input">${escapeHTML(initialProblem?.sample_input || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Sample Output</label>
        <textarea class="text-area" name="sample_output">${escapeHTML(initialProblem?.sample_output || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Hint</label>
        <textarea class="text-area" name="hint">${escapeHTML(initialProblem?.hint || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Source</label>
        <input class="text-input" name="source" value="${escapeHTML(initialProblem?.source || "SEU OJ")}" />
      </div>
      <div>
        <label class="field-label">Judge Mode</label>
        <select class="select-input" name="judge_mode">
          <option value="standard" ${initialProblem?.judge_mode === "standard" || !initialProblem ? "selected" : ""}>standard</option>
        </select>
      </div>
      <div>
        <label class="field-label">Time Limit (ms)</label>
        <input class="text-input" name="time_limit_ms" type="number" value="${escapeHTML(initialProblem?.time_limit_ms || 1000)}" min="1" required />
      </div>
      <div>
        <label class="field-label">Memory Limit (MB)</label>
        <input class="text-input" name="memory_limit_mb" type="number" value="${escapeHTML(initialProblem?.memory_limit_mb || 128)}" min="1" required />
      </div>
      <div class="full">
        <label class="field-label"><input type="checkbox" name="visible" ${initialProblem?.visible ?? true ? "checked" : ""} /> Visible</label>
      </div>
      <div class="full detail-card">
        <div class="view-header" style="margin-bottom:12px;">
          <div>
            <h3>Sample Testcases</h3>
            <p class="view-subtitle">Shown to users on the problem page.</p>
          </div>
          <button class="ghost-button" type="button" id="add-sample-case">Add Sample</button>
        </div>
        <div id="sample-cases"></div>
      </div>
      <div class="full detail-card">
        <div class="view-header" style="margin-bottom:12px;">
          <div>
            <h3>Hidden Testcases</h3>
            <p class="view-subtitle">Used by judge worker only.</p>
          </div>
          <button class="ghost-button" type="button" id="add-hidden-case">Add Hidden</button>
        </div>
        <div id="hidden-cases"></div>
      </div>
      <div class="full">
        <button class="primary-button" type="submit">${problemID ? "Update Problem" : "Create Problem"}</button>
      </div>
    </form>
  `;

  const sampleCases = document.getElementById("sample-cases");
  const hiddenCases = document.getElementById("hidden-cases");
  let sampleIndex = 0;
  let hiddenIndex = 0;

  const addCaseCard = (container, type, index, defaults = {}) => {
    const wrapper = document.createElement("div");
    wrapper.className = "detail-card";
    wrapper.style.marginBottom = "12px";
    wrapper.dataset.caseType = type;
    wrapper.dataset.caseIndex = String(index);
    wrapper.innerHTML = `
      <div class="view-header" style="margin-bottom:12px;">
        <div>
          <h3>${type === "sample" ? "Sample" : "Hidden"} #${index + 1}</h3>
          <p class="view-subtitle">${type === "sample" ? "Visible testcase" : "Judge-only testcase"}</p>
        </div>
        <button class="ghost-button remove-case" type="button">Remove</button>
      </div>
      <div class="grid-form">
        <div>
          <label class="field-label">Input</label>
          <textarea class="text-area" data-field="input">${escapeHTML(defaults.input || "")}</textarea>
        </div>
        <div>
          <label class="field-label">Output</label>
          <textarea class="text-area" data-field="output">${escapeHTML(defaults.output || "")}</textarea>
        </div>
        <div>
          <label class="field-label">Score</label>
          <input class="text-input" data-field="score" type="number" min="0" value="${escapeHTML(defaults.score ?? (type === "hidden" ? 100 : 0))}" />
        </div>
        <div>
          <label class="field-label">Sort Order</label>
          <input class="text-input" data-field="sort_order" type="number" min="1" value="${escapeHTML(defaults.sort_order ?? (index + 1))}" />
        </div>
        <div class="full">
          <label class="field-label"><input data-field="is_active" type="checkbox" ${defaults.is_active ?? true ? "checked" : ""} /> Active</label>
        </div>
      </div>
    `;
    wrapper.querySelector(".remove-case").addEventListener("click", () => {
      wrapper.remove();
      renumberCases(container, type);
    });
    container.appendChild(wrapper);
    renumberCases(container, type);
  };

  const renumberCases = (container, type) => {
    [...container.children].forEach((card, idx) => {
      card.dataset.caseIndex = String(idx);
      const title = card.querySelector("h3");
      if (title) {
        title.textContent = `${type === "sample" ? "Sample" : "Hidden"} #${idx + 1}`;
      }
    });
  };

  document.getElementById("add-sample-case").addEventListener("click", () => {
    addCaseCard(sampleCases, "sample", sampleIndex++);
  });
  document.getElementById("add-hidden-case").addEventListener("click", () => {
    addCaseCard(hiddenCases, "hidden", hiddenIndex++);
  });

  if (sampleCasesInitial.length) {
    sampleCasesInitial.forEach((item) => addCaseCard(sampleCases, "sample", sampleIndex++, {
      input: item.input_data,
      output: item.output_data,
      score: item.score,
      sort_order: item.sort_order,
      is_active: item.is_active,
    }));
  } else {
    addCaseCard(sampleCases, "sample", sampleIndex++, { input: "1 2", output: "3", score: 0, sort_order: 1, is_active: true });
  }

  if (hiddenCasesInitial.length) {
    hiddenCasesInitial.forEach((item) => addCaseCard(hiddenCases, "hidden", hiddenIndex++, {
      input: item.input_data,
      output: item.output_data,
      score: item.score,
      sort_order: item.sort_order,
      is_active: item.is_active,
    }));
  } else {
    addCaseCard(hiddenCases, "hidden", hiddenIndex++, { input: "5 7", output: "12", score: 100, sort_order: 2, is_active: true });
  }

  document.getElementById("admin-problem-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);

    const testcases = [];
    [...sampleCases.children].forEach((card, index) => {
      const input = card.querySelector('[data-field="input"]').value;
      const output = card.querySelector('[data-field="output"]').value;
      const score = Number(card.querySelector('[data-field="score"]').value || 0);
      const sortOrder = Number(card.querySelector('[data-field="sort_order"]').value || (index + 1));
      const isActive = card.querySelector('[data-field="is_active"]').checked;
      if (input.trim() || output.trim()) {
        testcases.push({
          case_type: "sample",
          input_data: input,
          output_data: output,
          score,
          sort_order: sortOrder,
          is_active: isActive,
        });
      }
    });

    [...hiddenCases.children].forEach((card, index) => {
      const input = card.querySelector('[data-field="input"]').value;
      const output = card.querySelector('[data-field="output"]').value;
      const score = Number(card.querySelector('[data-field="score"]').value || 100);
      const sortOrder = Number(card.querySelector('[data-field="sort_order"]').value || (sampleCases.children.length + index + 1));
      const isActive = card.querySelector('[data-field="is_active"]').checked;
      if (input.trim() && output.trim()) {
        testcases.push({
          case_type: "hidden",
          input_data: input,
          output_data: output,
          score,
          sort_order: sortOrder,
          is_active: isActive,
        });
      }
    });

    if (!testcases.some((item) => item.case_type === "hidden")) {
      setFlash("At least one hidden testcase is required", true);
      return;
    }

    const payload = {
      display_id: (form.get("display_id") || "").toString(),
      title: (form.get("title") || "").toString(),
      description: (form.get("description") || "").toString(),
      input_desc: (form.get("input_desc") || "").toString(),
      output_desc: (form.get("output_desc") || "").toString(),
      sample_input: (form.get("sample_input") || "").toString(),
      sample_output: (form.get("sample_output") || "").toString(),
      hint: (form.get("hint") || "").toString(),
      source: (form.get("source") || "").toString(),
      judge_mode: (form.get("judge_mode") || "").toString(),
      time_limit_ms: Number(form.get("time_limit_ms") || 1000),
      memory_limit_mb: Number(form.get("memory_limit_mb") || 128),
      visible: form.get("visible") === "on",
      testcases,
    };

    try {
      const result = await apiFetch(problemID ? `/admin/problems/${problemID}` : "/admin/problems", {
        method: problemID ? "PUT" : "POST",
        body: JSON.stringify(payload),
      });
      setFlash(`Problem ${problemID ? "updated" : "created"}: #${result.problem_id}`, false);
      location.hash = `#/problems/${result.problem_id}`;
    } catch (err) {
      setFlash(err.message, true);
    }
  });
}

async function renderAdminProblems() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before managing problems.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access problem management.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading admin problems...</p></div>`;
  try {
    const queryString = getCurrentHashPath().split("?")[1] || "";
    const params = new URLSearchParams(queryString);
    const keyword = params.get("keyword") || "";
    const includeHidden = params.get("include_hidden") === "true";
    const page = params.get("page") || "1";
    const pageSize = params.get("page_size") || "100";

    const query = new URLSearchParams({
      page,
      page_size: pageSize,
      include_hidden: String(includeHidden),
    });
    if (keyword) {
      query.set("keyword", keyword);
    }

    const data = await apiFetch(`/admin/problems?${query.toString()}`, { method: "GET" });
    const list = data.list || [];
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Manage Problems</h1>
          <p class="view-subtitle">Current lightweight admin console. Create new problems and inspect existing visible ones.</p>
        </div>
        <div>
          <a class="primary-button" href="#/admin/problems/new">New Problem</a>
        </div>
      </div>
      <form id="admin-problem-filters" class="toolbar">
        <div>
          <label class="field-label">Keyword</label>
          <input class="text-input" name="keyword" value="${escapeHTML(keyword)}" placeholder="Search title" />
        </div>
        <div>
          <label class="field-label">Include Hidden</label>
          <select class="select-input" name="include_hidden">
            <option value="false" ${!includeHidden ? "selected" : ""}>Visible only</option>
            <option value="true" ${includeHidden ? "selected" : ""}>Visible + Hidden</option>
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
          <button class="ghost-button" type="button" id="clear-admin-problem-filters">Clear</button>
        </div>
      </form>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Display</th>
            <th>Title</th>
            <th>Mode</th>
            <th>Visible</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          ${list.map((item) => `
            <tr>
              <td>${item.id}</td>
              <td class="mono">${escapeHTML(item.display_id || "-")}</td>
              <td>${escapeHTML(item.title)}</td>
              <td><span class="status-pill status-neutral">${escapeHTML(item.judge_mode)}</span></td>
              <td><span class="status-pill ${item.visible ? "status-accepted" : "status-hidden"}">${item.visible ? "Visible" : "Hidden"}</span></td>
              <td class="mono">${escapeHTML(item.created_at)}</td>
              <td>
                <a class="table-link" href="#/admin/problems/${item.id}">View</a>
                <span class="mono"> | </span>
                <a class="table-link" href="#/admin/problems/${item.id}/edit">Edit</a>
                <span class="mono"> | </span>
                <button class="link-button admin-delete-problem" data-problem-id="${item.id}">Delete</button>
              </td>
            </tr>
          `).join("")}
        </tbody>
      </table>
      <div class="detail-card" style="margin-top:16px;">
        <h3>Note</h3>
        <p>This page now uses admin APIs. You can filter by title, include hidden problems, and jump directly to edit/delete actions.</p>
      </div>
    `;

    document.getElementById("admin-problem-filters").addEventListener("submit", (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const next = new URLSearchParams();
      const nextKeyword = (form.get("keyword") || "").toString().trim();
      const nextIncludeHidden = (form.get("include_hidden") || "false").toString();
      const nextPage = (form.get("page") || "1").toString().trim() || "1";
      const nextPageSize = (form.get("page_size") || "100").toString().trim() || "100";

      next.set("page", nextPage);
      next.set("page_size", nextPageSize);
      next.set("include_hidden", nextIncludeHidden);
      if (nextKeyword) {
        next.set("keyword", nextKeyword);
      }

      location.hash = `#/admin/problems?${next.toString()}`;
    });

    document.getElementById("clear-admin-problem-filters").addEventListener("click", () => {
      location.hash = "#/admin/problems";
    });

    document.querySelectorAll(".admin-delete-problem").forEach((button) => {
      button.addEventListener("click", async () => {
        const id = button.dataset.problemId;
        const confirmed = window.confirm(`Delete problem #${id}? This will also remove its testcases.`);
        if (!confirmed) {
          return;
        }

        try {
          await apiFetch(`/admin/problems/${id}`, { method: "DELETE" });
          setFlash(`Problem #${id} deleted`, false);
          await renderAdminProblems();
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load admin problems failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function renderAdminCaseTable(list) {
  return `
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Type</th>
          <th>Input</th>
          <th>Output</th>
          <th>Score</th>
          <th>Order</th>
          <th>Active</th>
        </tr>
      </thead>
      <tbody>
        ${list.map((item) => `
          <tr>
            <td>${item.id}</td>
            <td>${escapeHTML(item.case_type)}</td>
            <td><pre style="margin:0; white-space:pre-wrap;">${escapeHTML(item.input_data || "")}</pre></td>
            <td><pre style="margin:0; white-space:pre-wrap;">${escapeHTML(item.output_data || "")}</pre></td>
            <td>${item.score}</td>
            <td>${item.sort_order}</td>
            <td>${item.is_active ? "Yes" : "No"}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;
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
        <p class="view-subtitle">Problem ${detail.problem_id} / ${escapeHTML(detail.language)}</p>
      </div>
      <div style="display:flex; gap:10px;">
        <button class="ghost-button" id="reuse-code-btn">Reuse Code</button>
        <a class="ghost-button" href="#/problems/${detail.problem_id}">Back to Problem</a>
        <a class="ghost-button" href="#/submissions">Back</a>
      </div>
    </div>
    <section class="verdict-banner ${verdictTone}">
      <div>
        <h2 class="verdict-title">${escapeHTML(detail.status)}</h2>
        <p class="verdict-subtitle">Submission for problem ${detail.problem_id}, created at ${escapeHTML(detail.created_at)}</p>
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
    location.hash = `#/problems/${detail.problem_id}`;
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
