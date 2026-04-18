const state = {
  apiBase: localStorage.getItem("seuoj_api_base") || `${location.origin}/api`,
  token: localStorage.getItem("seuoj_token") || "",
  user: readJSON("seuoj_user"),
  problems: [],
  problemDetail: null,
  submissions: [],
  submissionDetail: null,
  runResult: null,
  runResultPending: false,
  submissionPollTimer: null,
  submissionsPollTimer: null,
  problemTitleMap: {},
  workbenchLeftWidth: Number(localStorage.getItem("seuoj_workbench_left_width")) || 48,
  runResultHeight: Number(localStorage.getItem("seuoj_run_result_height")) || 180,
  contestPollTimer: null,
  runResultUIAbort: null,
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

async function bootstrapApp() {
  try {
    renderAccount();
    updateNavActive();
    if (state.token && !state.user) {
      await refreshCurrentUser();
    }
    if (!location.hash) {
      location.hash = "#/problems";
    }
    await renderRoute();
  } catch (err) {
    renderFatalError(err, "startup");
  }
}

async function renderRouteSafely() {
  try {
    await renderRoute();
  } catch (err) {
    renderFatalError(err, "route");
  }
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
    const message = data.message || "request failed";
    if (message === "invalid or expired token" || message === "missing authorization header") {
      persistAuth("", null);
      if (getCurrentHashPath() !== "/auth") {
        location.hash = "#/auth";
      }
      setFlash("Login expired, please login again", true);
    }
    throw new Error(message);
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

function renderFatalError(err, source = "unknown") {
  const message = err instanceof Error ? err.message : String(err || "Unknown error");
  console.error(`[frontend-${source}]`, err);
  setFlash(`Frontend ${source} error: ${message}`, true);
  if (!app) {
    return;
  }
  app.innerHTML = `
    <section class="detail-card">
      <h2>Frontend Render Error</h2>
      <p>The page failed to render normally.</p>
      <pre class="mono">${escapeHTML(message)}</pre>
      <div style="margin-top:14px; display:flex; gap:10px; flex-wrap:wrap;">
        <a class="ghost-button" href="#/problems">Go to Problems</a>
        <a class="ghost-button" href="#/auth">Go to Account</a>
        <button class="primary-button" type="button" id="reload-page-btn">Reload</button>
      </div>
    </section>
  `;

  const reloadButton = document.getElementById("reload-page-btn");
  if (reloadButton) {
    reloadButton.addEventListener("click", () => location.reload());
  }
}

function renderAccount() {
  const profileLink = document.getElementById("profile-link");
  const classesLink = document.getElementById("classes-link");

  if (profileLink) profileLink.style.display = state.user ? "block" : "none";
  if (classesLink) classesLink.style.display = state.user ? "block" : "none";

  if (!state.user) {
    accountSlot.innerHTML = `
      <span class="mono account-guest-label">guest</span>
      <a class="ghost-button account-login-button" href="#/auth">Login</a>
    `;
    return;
  }

  const teacherLinks = isTeacherUser()
    ? `
      <div class="account-menu-section">
        <span class="account-menu-label">Teaching</span>
        <a href="#/teacher/playlists">Teach Playlists</a>
        <a href="#/teacher/classes">Teach Classes</a>
      </div>
    `
    : "";

  const adminLinks = state.user.role === "admin"
    ? `
      <div class="account-menu-section">
        <span class="account-menu-label">Admin</span>
        <a href="#/admin/problems">Manage Problems</a>
        <a href="#/admin/contests">Manage Contests</a>
        <a href="#/admin/submissions">Manage Submissions</a>
        <a href="#/admin/users">Manage Users</a>
        <a href="#/admin/announcements">Manage Announcements</a>
        <a href="#/admin/problems/new">Create Problem</a>
      </div>
    `
    : "";

  accountSlot.innerHTML = `
    <details class="account-menu">
      <summary>
        <span class="mono">${escapeHTML(state.user.username)}</span>
        <span class="status-pill ${state.user.role === "admin" ? "status-accepted" : state.user.role === "teacher" ? "status-neutral" : "status-pending"}">${escapeHTML(state.user.role)}</span>
      </summary>
      <div class="account-menu-panel">
        <div class="account-menu-section">
          <span class="account-menu-label">Workspace</span>
          <a href="#/profile">Profile</a>
          <a href="#/classes">Classes</a>
          <a href="#/submissions">Submissions</a>
        </div>
        ${teacherLinks}
        ${adminLinks}
        <div class="account-menu-section danger-zone">
          <button class="ghost-button" id="logout-btn" type="button">Logout</button>
        </div>
      </div>
    </details>
  `;

  document.getElementById("logout-btn").addEventListener("click", () => {
    persistAuth("", null);
    setFlash("Logged out", false);
    location.hash = "#/auth";
  });
}
function updateNavActive() {
  const routePath = getCurrentHashPath().split("?")[0] || "/";
  document.querySelectorAll(".topnav a").forEach((link) => {
    const href = link.getAttribute("href") || "";
    const target = href.replace(/^#/, "") || "/";
    let active = false;
    if (target === "/") {
      active = routePath === "/";
    } else if (routePath === target) {
      active = true;
    } else if (routePath.startsWith(target + "/")) {
      active = true;
    }
    link.classList.toggle("active", active);
  });
}

function getCurrentHashPath() {
  return location.hash.replace(/^#/, "") || "/";
}

function getHashQueryParams() {
  const hash = getCurrentHashPath();
  const query = hash.includes("?") ? hash.split("?")[1] : "";
  return new URLSearchParams(query);
}

async function renderRoute() {
  stopSubmissionPolling();
  stopSubmissionsPolling();
  stopContestPolling();
  setFlash("");
  updateNavActive();

  const hash = getCurrentHashPath();
  const routePath = hash.split("?")[0] || "/";
  const contestProblemMatch = routePath.match(/^\/contests\/(\d+)\/problems\/(\d+)$/);
  const adminContestAnnouncementEditMatch = routePath.match(/^\/admin\/contests\/(\d+)\/announcements\/(\d+)\/edit$/);
  const adminContestAnnouncementNewMatch = routePath.match(/^\/admin\/contests\/(\d+)\/announcements\/new$/);
  const adminContestEditMatch = routePath.match(/^\/admin\/contests\/(\d+)\/edit$/);
  const adminContestDetailMatch = routePath.match(/^\/admin\/contests\/(\d+)$/);
  document.body.classList.toggle("problem-fullscreen", routePath.startsWith("/problems/") || !!contestProblemMatch);

  if (routePath === "/") return renderHome();
  if (routePath === "/playlists") return renderPlaylists();
  if (routePath.startsWith("/playlists/")) return renderPlaylistDetail(routePath.split("/")[2]);
  if (routePath === "/classes") return renderClasses();
  if (routePath.startsWith("/classes/")) return renderClassDetail(routePath.split("/")[2]);
  if (routePath.startsWith("/assignments/")) return renderAssignmentDetail(routePath.split("/")[2]);
  if (routePath === "/teacher/playlists") return renderTeacherPlaylists();
  if (routePath.startsWith("/teacher/playlists/")) return renderTeacherPlaylistDetail(routePath.split("/")[3]);
  if (routePath === "/teacher/classes") return renderTeacherClasses();
  if (routePath.startsWith("/teacher/classes/")) return renderTeacherClassDetail(routePath.split("/")[3]);
  if (routePath.startsWith("/teacher/assignments/")) return renderTeacherAssignmentOverview(routePath.split("/")[3]);
  if (routePath === "/rankings") return renderRankings();
  if (routePath.startsWith("/forum/topics/")) return renderForumTopicDetail(routePath.split("/")[3]);
  if (routePath === "/forum") return renderForum();
  if (routePath.startsWith("/contests/") && contestProblemMatch) return renderContestProblemDetail(contestProblemMatch[1], contestProblemMatch[2]);
  if (routePath.startsWith("/contests/") && !routePath.includes("/problems/")) return renderContestDetail(routePath.split("/")[2]);
  if (routePath.startsWith("/contests")) return renderContests();
  if (routePath === "/announcements") return renderAnnouncements();
  if (routePath.startsWith("/announcements/")) return renderAnnouncementDetail(routePath.split("/")[2]);
  if (routePath.startsWith("/problems/")) return renderProblemDetail(routePath.split("/")[2]);
  if (routePath.startsWith("/problems")) return renderProblems();
  if (routePath === "/submissions") return renderMySubmissions();
  if (routePath.startsWith("/submissions/")) return renderSubmissionDetail(routePath.split("/")[2]);
  if (routePath === "/profile") return renderProfile();
  if (routePath === "/admin/submissions") return renderAdminSubmissions();
  if (routePath === "/admin/users") return renderAdminUsers();
  if (routePath === "/admin/announcements") return renderAdminAnnouncements();
  if (routePath.startsWith("/admin/announcements/") && routePath.endsWith("/edit")) return renderAdminAnnouncementEdit(routePath.split("/")[3]);
  if (routePath === "/admin/announcements/new") return renderAdminAnnouncementCreate();
  if (routePath === "/admin/contests") return renderAdminContests();
  if (routePath === "/admin/contests/new") return renderAdminContestCreate();
  if (adminContestAnnouncementNewMatch) return renderAdminContestAnnouncementCreate(adminContestAnnouncementNewMatch[1]);
  if (adminContestAnnouncementEditMatch) return renderAdminContestAnnouncementEdit(adminContestAnnouncementEditMatch[1], adminContestAnnouncementEditMatch[2]);
  if (adminContestEditMatch) return renderAdminContestEdit(adminContestEditMatch[1]);
  if (adminContestDetailMatch) return renderAdminContestDetail(adminContestDetailMatch[1]);
  if (routePath === "/admin/problems") return renderAdminProblems();
  if (routePath === "/admin/problems/new") return renderAdminProblemCreate();
  if (routePath.startsWith("/admin/problems/") && routePath.endsWith("/edit")) return renderAdminProblemEdit(routePath.split("/")[3]);
  if (routePath.startsWith("/admin/problems/") && !routePath.endsWith("/edit")) return renderAdminProblemDetail(routePath.split("/")[3]);
  if (routePath === "/auth") return renderAuth();
  return renderAuth();
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
  let overviewStats = null;
  let myStats = null;
  let adminStats = null;
  let contests = [];

  try {
    const [problemsData, overviewData, contestData] = await Promise.all([
      apiFetch("/problems?page=1&page_size=5", { method: "GET" }),
      apiFetch("/stats/overview", { method: "GET" }).catch(() => null),
      apiFetch("/contests?page=1&page_size=6", { method: "GET" }).catch(() => null),
    ]);
    latestProblems = problemsData.list || [];
    totalProblems = problemsData.total || latestProblems.length;
    overviewStats = overviewData;
    contests = contestData?.list || [];
  } catch {
    latestProblems = [];
    totalProblems = 0;
    contests = [];
  }

  if (state.token) {
    try {
      const [submissionsData, myStatsData, adminStatsData] = await Promise.all([
        apiFetch("/submissions/my?page=1&page_size=5", { method: "GET" }),
        apiFetch("/stats/me", { method: "GET" }).catch(() => null),
        state.user?.role === "admin" ? apiFetch("/stats/admin", { method: "GET" }).catch(() => null) : Promise.resolve(null),
      ]);
      latestSubmissions = submissionsData.list || [];
      myStats = myStatsData;
      adminStats = adminStatsData;
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
    <section class="detail-card" style="margin-top:18px;">
      <div class="verdict-summary-grid">
        <div class="verdict-summary-card"><span class="status-pill status-neutral">Problems</span><strong>${overviewStats?.problems_total ?? totalProblems}</strong></div>
        <div class="verdict-summary-card"><span class="status-pill status-accepted">Accepted Submissions</span><strong>${overviewStats?.accepted_submissions ?? 0}</strong></div>
        <div class="verdict-summary-card"><span class="status-pill status-neutral">Users</span><strong>${overviewStats?.users_total ?? 0}</strong></div>
        <div class="verdict-summary-card"><span class="status-pill status-pending">My Solved</span><strong>${myStats?.accepted_problems ?? 0}</strong></div>
        ${adminStats ? `<div class="verdict-summary-card"><span class="status-pill status-pending">Queue</span><strong>${adminStats.queue_length}</strong></div>` : ""}
        ${adminStats ? `<div class="verdict-summary-card"><span class="status-pill status-error">System Errors</span><strong>${adminStats.system_errors}</strong></div>` : ""}
        <div class="verdict-summary-card"><span class="status-pill status-neutral">Contests</span><strong>${contests.length}</strong></div>
      </div>
    </section>
    <section class="detail-card" style="margin-top:18px;">
      <div class="view-header">
        <div>
          <h3>Contest Snapshot</h3>
          <p class="view-subtitle">Recent public contests and their current phase.</p>
        </div>
        <a class="ghost-button" href="#/contests">All Contests</a>
      </div>
      ${renderContestStatusBreakdown(contests)}
      <div style="margin-top:16px;">${renderContestDeck(contests.slice(0, 3))}</div>
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

async function renderProfile() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login to view your profile.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading profile...</p></div>`;
  try {
    const [me, stats] = await Promise.all([
      apiFetch("/auth/me", { method: "GET" }),
      apiFetch("/stats/me", { method: "GET" }),
    ]);

    persistAuth(state.token, me);

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Profile</h1>
          <p class="view-subtitle">Manage your account information and inspect your personal judge footprint.</p>
        </div>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          <h3>Profile</h3>
          <form id="profile-form" class="toolbar">
            <div>
              <label class="field-label">Username</label>
              <input class="text-input" name="username" value="${escapeHTML(me.username)}" required />
            </div>
            <div>
              <label class="field-label">Student ID</label>
              <input class="text-input" name="userid" value="${escapeHTML(me.userid)}" required />
            </div>
            <div>
              <label class="field-label">Role</label>
              <input class="text-input" value="${escapeHTML(me.role)}" disabled />
            </div>
            <div>
              <label class="field-label">Status</label>
              <input class="text-input" value="${escapeHTML(me.status)}" disabled />
            </div>
            <div class="full" style="display:flex; gap:10px; align-items:center;">
              <button class="primary-button" type="submit">Save Profile</button>
            </div>
          </form>
        </article>
        <aside class="detail-card">
          <h3>Password</h3>
          <form id="password-form" class="toolbar">
            <div class="full">
              <label class="field-label">Current Password</label>
              <input class="text-input" name="current_password" type="password" required />
            </div>
            <div class="full">
              <label class="field-label">New Password</label>
              <input class="text-input" name="new_password" type="password" required />
            </div>
            <div class="full" style="display:flex; gap:10px; align-items:center;">
              <button class="ghost-button" type="submit">Change Password</button>
            </div>
          </form>
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header">
          <div>
            <h3>My Stats</h3>
            <p class="view-subtitle">Aggregated from your submissions and accepted problems.</p>
          </div>
        </div>
        <div class="verdict-summary-grid">
          <div class="verdict-summary-card"><span class="status-pill status-neutral">Submissions</span><strong>${stats.submissions_total}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-accepted">Accepted</span><strong>${stats.accepted_submissions}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-neutral">Solved Problems</span><strong>${stats.accepted_problems}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-pending">Pending / Running</span><strong>${stats.pending_submissions + stats.running_submissions}</strong></div>
        </div>
      </section>
      <section class="detail-grid" style="margin-top:18px;">
        <article class="detail-card">
          <h3>Status Breakdown</h3>
          ${renderCountTable(stats.status_breakdown || [], "Status", "Count")}
        </article>
        <article class="detail-card">
          <h3>Language Breakdown</h3>
          ${renderCountTable(stats.language_breakdown || [], "Language", "Count")}
        </article>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <h3>Recent Activity</h3>
        ${renderRecentActivityTable(stats.recent_activity || [])}
      </section>
    `;

    document.getElementById("profile-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        const updated = await apiFetch("/auth/profile", {
          method: "PUT",
          body: JSON.stringify({
            username: form.get("username"),
            userid: form.get("userid"),
          }),
        });
        persistAuth(state.token, updated);
        setFlash("Profile updated", false);
        await renderProfile();
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("password-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch("/auth/password", {
          method: "PUT",
          body: JSON.stringify({
            current_password: form.get("current_password"),
            new_password: form.get("new_password"),
          }),
        });
        setFlash("Password changed", false);
        event.currentTarget.reset();
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load profile failed: ${escapeHTML(err.message)}</p></div>`;
  }
}




