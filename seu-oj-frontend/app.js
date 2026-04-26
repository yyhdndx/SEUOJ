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
  problemCodeEditor: null,
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
    initNavMenus();
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

function initNavMenus() {
  document.addEventListener("click", (event) => {
    const target = event.target;
    if (!(target instanceof Element)) {
      return;
    }

    const clickedMenu = target.closest(".nav-menu");
    if (!clickedMenu) {
      closeNavMenus();
      return;
    }

    if (target.closest(".nav-menu-panel a")) {
      closeNavMenus();
      return;
    }

    closeNavMenus(clickedMenu);
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      closeNavMenus();
    }
  });
}

function closeNavMenus(exceptMenu = null) {
  document.querySelectorAll(".nav-menu[open]").forEach((menu) => {
    if (menu !== exceptMenu) {
      menu.open = false;
    }
  });
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
  if (state.problemCodeEditor) {
    state.problemCodeEditor.destroy();
    state.problemCodeEditor = null;
  }
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
  document.body.classList.toggle("auth-page", routePath === "/auth");

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
    <div style="margin:0 auto; max-width:460px;">
      <form id="login-form" class="detail-card">
        <h3>Login</h3>
        <label class="field-label">Username</label>
        <input class="text-input" name="username" required />
        <label class="field-label">Password</label>
        <input class="text-input" name="password" type="password" required />
        <div style="margin-top:14px;">
          <button class="primary-button" type="submit">Login</button>
        </div>
        <p class="view-subtitle" style="margin:14px 0 0;">
          No account yet?
          <button class="link-button" id="show-register-btn" type="button">Register</button>
        </p>
      </form>
      <form id="register-form" class="detail-card hidden">
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
        <p class="view-subtitle" style="margin:14px 0 0;">
          Already have an account?
          <button class="link-button" id="show-login-btn" type="button">Login</button>
        </p>
      </form>
    </div>
  `;

  const loginForm = document.getElementById("login-form");
  const registerForm = document.getElementById("register-form");
  document.getElementById("show-register-btn").addEventListener("click", () => {
    loginForm.classList.add("hidden");
    registerForm.classList.remove("hidden");
  });
  document.getElementById("show-login-btn").addEventListener("click", () => {
    registerForm.classList.add("hidden");
    loginForm.classList.remove("hidden");
  });

  loginForm.addEventListener("submit", async (event) => {
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

  registerForm.addEventListener("submit", async (event) => {
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
      registerForm.classList.add("hidden");
      loginForm.classList.remove("hidden");
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
  let announcements = [];
  let forumTopics = [];

  try {
    const [problemsData, overviewData, contestData, announcementsData, forumData] = await Promise.all([
      apiFetch("/problems?page=1&page_size=5", { method: "GET" }),
      apiFetch("/stats/overview", { method: "GET" }).catch(() => null),
      apiFetch("/contests?page=1&page_size=6", { method: "GET" }).catch(() => null),
      apiFetch("/announcements?page=1&page_size=3", { method: "GET" }).catch(() => null),
      apiFetch("/forum/topics?page=1&page_size=3", { method: "GET" }).catch(() => null),
    ]);
    latestProblems = problemsData.list || [];
    totalProblems = problemsData.total || latestProblems.length;
    overviewStats = overviewData;
    contests = contestData?.list || [];
    announcements = announcementsData?.list || [];
    forumTopics = forumData?.list || [];
  } catch {
    latestProblems = [];
    totalProblems = 0;
    contests = [];
    announcements = [];
    forumTopics = [];
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

  const userRole = state.user?.role || "guest";
  const isAdmin = userRole === "admin";
  const isTeacher = state.user && isTeacherUser();
  const showTeachingSnapshot = userRole === "teacher";
  const primaryAction = getHomePrimaryAction(latestSubmission);
  const quickActions = getHomeQuickActions(userRole, latestSubmission);
  const runningContest = contests.find((item) => item.status === "running");
  const nextContest = runningContest || contests.find((item) => item.status === "upcoming") || contests[0] || null;

  app.innerHTML = `
    <section class="dashboard-hero">
      <div>
        <div class="pill-row">
          <span class="status-pill ${isAdmin ? "status-accepted" : isTeacher ? "status-pending" : "status-neutral"}">${escapeHTML(getHomeRoleLabel(userRole))}</span>
          <span class="status-pill ${state.token ? "status-accepted" : "status-neutral"}">${state.token ? "Signed in" : "Guest"}</span>
        </div>
        <h1 class="dashboard-title">${escapeHTML(getHomeGreeting())}</h1>
        <p class="view-subtitle">${escapeHTML(getHomeRoleHint(userRole))}</p>
      </div>
      <div class="dashboard-hero-action">
        <span class="metric-label">Next best action</span>
        <a class="primary-button" href="${primaryAction.href}" ${primaryAction.id ? `id="${primaryAction.id}"` : ""}>${escapeHTML(primaryAction.label)}</a>
        <p class="view-subtitle">${escapeHTML(primaryAction.hint)}</p>
      </div>
    </section>

    <section class="dashboard-section">
      <div class="view-header">
        <div>
          <h3>Quick Actions</h3>
          <p class="view-subtitle">Role-focused entry points for the work you are most likely to do now.</p>
        </div>
      </div>
      <div class="dashboard-action-grid">
        ${quickActions.map((item) => `
          <a class="dashboard-action-card ${item.primary ? "primary" : ""}" href="${item.href}">
            <span class="dashboard-action-label">${escapeHTML(item.label)}</span>
            <strong>${escapeHTML(item.title)}</strong>
            <span class="view-subtitle">${escapeHTML(item.hint)}</span>
          </a>
        `).join("")}
      </div>
    </section>

    <section class="dashboard-main-grid">
      <article class="detail-card dashboard-focus-card">
        <div class="view-header">
          <div>
            <h3>Continue Work</h3>
            <p class="view-subtitle">Latest verdict, failure reason, and resume action in one place.</p>
          </div>
          ${latestSubmission ? `<a class="ghost-button" href="#/submissions/${latestSubmission.id}">Open Detail</a>` : ""}
        </div>
        ${renderHomeContinueWork(latestSubmission, latestSubmissionDetail, latestAcceptedSubmission)}
      </article>
      <aside class="detail-card dashboard-activity-card">
        <div class="view-header">
          <div>
            <h3>Recent Activity</h3>
            <p class="view-subtitle">Recent submissions are summarized here instead of repeated across cards.</p>
          </div>
          <a class="ghost-button" href="#/submissions">All Submissions</a>
        </div>
        ${renderHomeActivityList(latestSubmissions)}
      </aside>
    </section>

    <section class="dashboard-secondary-grid">
      <article class="detail-card">
        <div class="view-header">
          <div>
            <h3>${showTeachingSnapshot ? "Teaching Snapshot" : "Problem Snapshot"}</h3>
            <p class="view-subtitle">${showTeachingSnapshot ? "Teaching shortcuts stay visible without overwhelming student workflow." : "A compact way back into the problem set."}</p>
          </div>
          <a class="ghost-button" href="${showTeachingSnapshot ? "#/teacher/classes" : "#/problems"}">${showTeachingSnapshot ? "Teacher Console" : "All Problems"}</a>
        </div>
        ${showTeachingSnapshot ? renderHomeTeachingSnapshot() : renderHomeProblemList(latestProblems)}
      </article>
      <article class="detail-card">
        <div class="view-header">
          <div>
            <h3>Contest Snapshot</h3>
            <p class="view-subtitle">Only the nearest contest context is promoted here.</p>
          </div>
          <a class="ghost-button" href="#/contests">All Contests</a>
        </div>
        ${renderHomeContestSnapshot(nextContest, contests)}
      </article>
    </section>

    <section class="dashboard-bottom-grid">
      <article class="detail-card dashboard-muted-card">
        <div class="view-header">
          <div>
            <h3>System Snapshot</h3>
            <p class="view-subtitle">Secondary numbers stay below the main workflow.</p>
          </div>
        </div>
        <div class="verdict-summary-grid">
          <div class="verdict-summary-card"><span class="status-pill status-neutral">Problems</span><strong>${overviewStats?.problems_total ?? totalProblems}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-accepted">Accepted</span><strong>${overviewStats?.accepted_submissions ?? 0}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-neutral">Users</span><strong>${overviewStats?.users_total ?? 0}</strong></div>
          ${state.token ? `<div class="verdict-summary-card"><span class="status-pill status-pending">My Solved</span><strong>${myStats?.accepted_problems ?? 0}</strong></div>` : ""}
          ${adminStats ? `<div class="verdict-summary-card"><span class="status-pill status-pending">Queue</span><strong>${adminStats.queue_length}</strong></div>` : ""}
          ${adminStats ? `<div class="verdict-summary-card"><span class="status-pill status-error">System Errors</span><strong>${adminStats.system_errors}</strong></div>` : ""}
        </div>
      </article>
      <article class="detail-card dashboard-muted-card">
        <div class="view-header">
          <div>
            <h3>Updates</h3>
            <p class="view-subtitle">Announcements and forum are kept as lightweight pointers.</p>
          </div>
          <div class="pill-row">
            <a class="ghost-button" href="#/announcements">Announcements</a>
            <a class="ghost-button" href="#/forum">Forum</a>
          </div>
        </div>
        ${renderHomeUpdates(announcements, forumTopics)}
      </article>
    </section>
  `;

  if (latestSubmission) {
    const continueBtn = document.getElementById("continue-last-work-btn") || document.getElementById("hero-continue-work-btn");
    if (continueBtn) {
      continueBtn.addEventListener("click", (event) => {
        event.preventDefault();
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

function getHomeGreeting() {
  if (!state.user) {
    return "Welcome to SEU OJ";
  }
  return `Welcome back, ${state.user.username}`;
}

function getHomeRoleLabel(role) {
  if (role === "admin") return "Administrator";
  if (role === "teacher") return "Teacher";
  if (state.token) return "Student";
  return "Guest";
}

function getHomeRoleHint(role) {
  if (role === "admin") {
    return "Manage the judging system, contest operations, problem set, submissions, and announcements from one portal.";
  }
  if (role === "teacher") {
    return "Jump into classes, assignments, playlists, and the student-facing practice workflow.";
  }
  if (state.token) {
    return "Continue solving, check recent verdicts, and move quickly into problems, contests, and homework.";
  }
  return "Browse public problems and contests, then sign in when you are ready to submit and track progress.";
}

function getHomePrimaryAction(latestSubmission) {
  if (!state.token) {
    return { label: "Login or Register", href: "#/auth", hint: "Sign in to submit code and see your recent activity." };
  }
  if (latestSubmission) {
    return { label: isSubmissionPollingStatus(latestSubmission.status) ? "Track Latest Verdict" : "Continue Last Work", href: "#/submissions", id: "hero-continue-work-btn", hint: `Submission #${latestSubmission.id} is your freshest context.` };
  }
  if (state.user?.role === "admin") {
    return { label: "Manage Problems", href: "#/admin/problems", hint: "Start from the highest-impact admin workspace." };
  }
  if (state.user?.role === "teacher") {
    return { label: "Open Classes", href: "#/teacher/classes", hint: "Review class activity and publish assignments." };
  }
  return { label: "Browse Problems", href: "#/problems", hint: "Pick a problem and get to a verdict." };
}

function getHomeQuickActions(role, latestSubmission) {
  if (!state.token) {
    return [
      { label: "Account", title: "Login or Register", hint: "Unlock submissions and personal history.", href: "#/auth", primary: true },
      { label: "Practice", title: "Browse Problems", hint: "Explore visible problems before signing in.", href: "#/problems" },
      { label: "Events", title: "View Contests", hint: "Check public contest schedules.", href: "#/contests" },
    ];
  }

  if (role === "admin") {
    return [
      { label: "Admin", title: "Manage Problems", hint: "Create, edit, and review problem inventory.", href: "#/admin/problems", primary: true },
      { label: "Ops", title: "Manage Contests", hint: "Maintain contest windows and problem sets.", href: "#/admin/contests" },
      { label: "Judge", title: "Review Submissions", hint: "Inspect queue and recent judging output.", href: "#/admin/submissions" },
      { label: "Comms", title: "Announcements", hint: "Publish system-facing updates.", href: "#/admin/announcements" },
    ];
  }

  if (role === "teacher") {
    return [
      { label: "Teaching", title: "Classes", hint: "Open class rosters and assignment progress.", href: "#/teacher/classes", primary: true },
      { label: "Content", title: "Playlists", hint: "Curate practice sets and homework sources.", href: "#/teacher/playlists" },
      { label: "Practice", title: "Problem Set", hint: "Use student-facing practice flow.", href: "#/problems" },
      { label: "Verdicts", title: "My Submissions", hint: "Review your latest judge results.", href: "#/submissions" },
    ];
  }

  return [
    { label: "Practice", title: latestSubmission ? "Continue Work" : "Browse Problems", hint: latestSubmission ? "Return to your most recent submission context." : "Find the next problem to solve.", href: latestSubmission ? "#/submissions" : "#/problems", primary: true },
    { label: "Verdicts", title: "My Submissions", hint: "Check judge feedback and reuse code.", href: "#/submissions" },
    { label: "Events", title: "Contests", hint: "Join or practice contest problem sets.", href: "#/contests" },
    { label: "Learning", title: "Classes", hint: "Open assigned playlists and homework.", href: "#/classes" },
  ];
}

function renderHomeContinueWork(latestSubmission, latestSubmissionDetail, latestAcceptedSubmission) {
  if (!state.token) {
    return `<p class="view-subtitle">Login to see your latest verdict, failure reason, accepted record, and resume action.</p>`;
  }
  if (!latestSubmission) {
    return `
      <div class="dashboard-empty-state">
        <strong>No submission yet</strong>
        <p class="view-subtitle">Start from the problem set, then this area will become your resume point.</p>
        <a class="primary-button" href="#/problems">Browse Problems</a>
      </div>
    `;
  }

  return `
    <div class="verdict-banner ${getVerdictTone(latestSubmission.status)} dashboard-verdict">
      <div>
        <h2 class="verdict-title">${escapeHTML(latestSubmission.status)}</h2>
        <p class="verdict-subtitle">Problem ${latestSubmission.problem_id} / Submission #${latestSubmission.id}</p>
        ${renderLatestFailureSummary(latestSubmission, latestSubmissionDetail)}
      </div>
      <div class="verdict-stats">
        <div class="verdict-stat"><span class="verdict-stat-label">Passed</span><span class="verdict-stat-value">${latestSubmission.passed_count}/${latestSubmission.total_count}</span></div>
        <div class="verdict-stat"><span class="verdict-stat-label">Runtime</span><span class="verdict-stat-value">${latestSubmission.runtime_ms ?? "-"} ms</span></div>
        <div class="verdict-stat"><span class="verdict-stat-label">Created</span><span class="verdict-stat-value mono">${escapeHTML(latestSubmission.created_at)}</span></div>
        <div class="verdict-stat"><span class="verdict-stat-label">Action</span><span class="verdict-stat-value">${renderHomeActionForLatestSubmission(latestSubmission)}</span></div>
      </div>
    </div>
    ${latestAcceptedSubmission ? `
      <div class="dashboard-accepted-strip">
        <span class="status-pill status-accepted">Latest Accepted</span>
        <span>Problem ${latestAcceptedSubmission.problem_id} / Submission #${latestAcceptedSubmission.id}</span>
        <button class="ghost-button" id="open-accepted-problem-btn" type="button">Open Problem</button>
      </div>
    ` : ""}
  `;
}

function renderHomeActivityList(list) {
  if (!state.token) {
    return `<p class="view-subtitle">Please login to view recent submissions.</p>`;
  }
  if (!list.length) {
    return `<p class="view-subtitle">No submission yet.</p>`;
  }
  return `
    <div class="dashboard-activity-list">
      ${list.map((item) => `
        <a class="dashboard-activity-row" href="#/submissions/${item.id}">
          <span class="mono">#${item.id}</span>
          <span>Problem ${item.problem_id}</span>
          <span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span>
          <span class="mono">${item.passed_count}/${item.total_count}</span>
        </a>
      `).join("")}
    </div>
  `;
}

function renderHomeProblemList(problems) {
  if (!problems.length) {
    return `<p class="view-subtitle">No problem data available.</p>`;
  }
  return `
    <div class="dashboard-list">
      ${problems.slice(0, 4).map((item) => `
        <a class="dashboard-list-row" href="#/problems/${item.id}">
          <span class="mono">#${item.id}</span>
          <strong>${escapeHTML(item.title)}</strong>
          <span class="view-subtitle">${item.time_limit_ms} ms / ${item.memory_limit_mb} MB</span>
        </a>
      `).join("")}
    </div>
  `;
}

function renderHomeTeachingSnapshot() {
  return `
    <div class="dashboard-action-grid compact">
      <a class="dashboard-action-card" href="#/teacher/classes"><span class="dashboard-action-label">Classes</span><strong>Manage Classes</strong><span class="view-subtitle">Rosters, assignments, and analytics.</span></a>
      <a class="dashboard-action-card" href="#/teacher/playlists"><span class="dashboard-action-label">Playlists</span><strong>Build Practice Sets</strong><span class="view-subtitle">Curate reusable problem lists.</span></a>
    </div>
  `;
}

function renderHomeContestSnapshot(contest, contests) {
  if (!contest) {
    return `<p class="view-subtitle">No contest scheduled yet.</p>`;
  }
  const running = contests.filter((item) => item.status === "running").length;
  const upcoming = contests.filter((item) => item.status === "upcoming").length;
  return `
    <div class="dashboard-contest-feature">
      <span class="status-pill ${contestStatusClass(contest.status)}">${escapeHTML(contest.status)}</span>
      <h4>${escapeHTML(contest.title)}</h4>
      <p class="view-subtitle">${renderContestStatusHint(contest)}</p>
      <div class="dashboard-mini-metrics">
        <span>${contest.problem_count} problems</span>
        <span>${contest.registered_count} registrations</span>
        <span>${running} running / ${upcoming} upcoming</span>
      </div>
      <a class="ghost-button" href="#/contests/${contest.id}">Open Contest</a>
    </div>
  `;
}

function renderHomeUpdates(announcements, topics) {
  const announcementItems = announcements.slice(0, 2).map((item) => `
    <a class="dashboard-list-row" href="#/announcements/${item.id}">
      <span class="status-pill ${item.is_pinned ? "status-pending" : "status-neutral"}">${item.is_pinned ? "Pinned" : "News"}</span>
      <strong>${escapeHTML(item.title)}</strong>
      <span class="view-subtitle mono">${escapeHTML(item.created_at)}</span>
    </a>
  `).join("");
  const topicItems = topics.slice(0, 2).map((item) => `
    <a class="dashboard-list-row" href="#/forum/topics/${item.id}">
      <span class="status-pill status-neutral">Forum</span>
      <strong>${escapeHTML(item.title)}</strong>
      <span class="view-subtitle">${escapeHTML(item.content_preview || "")}</span>
    </a>
  `).join("");
  return `<div class="dashboard-list">${announcementItems}${topicItems || ""}${!announcementItems && !topicItems ? `<p class="view-subtitle">No updates yet.</p>` : ""}</div>`;
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
