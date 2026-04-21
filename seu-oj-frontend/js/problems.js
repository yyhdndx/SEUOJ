// Problem domain pages and helpers
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
    const [problem, problemStats] = await Promise.all([
      apiFetch(`/problems/${id}`, { method: "GET" }),
      apiFetch(`/problems/${id}/stats`, { method: "GET" }).catch(() => null),
    ]);
    state.problemDetail = problem;
    if (!state.runResult || state.runResult.problemID !== problem.id || state.runResult.contestID) {
      state.runResult = null;
    }
    state.runResultPending = false;
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
            ${renderProblemCodeBlock("Sample Input", problem.sample_input)}
            ${renderProblemCodeBlock("Sample Output", problem.sample_output)}
            ${sampleCases.length ? `
              <div class="detail-block">
                <h3>Sample Testcases</h3>
                <p class="view-subtitle">${sampleCases.length} sample case(s) are available.</p>
                ${renderSampleCaseList(sampleCases)}
              </div>
            ` : ""}
            ${renderProblemStats(problemStats)}
            ${renderProblemBlock("Source / Hint", `${problem.source || ""}\n${problem.hint || ""}`.trim())}
            ${renderProblemSolutions(problem.solutions, problem.id)}
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
              <div id="run-result-slot">${renderRunResultPanel()}</div>
            </div>
          </form>
          </div>
        </aside>
      </section>
    `;

    initProblemWorkbenchUI();
    initRunResultUI();

    const codeEditor = document.getElementById("problem-code-editor");
    await mountProblemCodeEditor(codeEditor);
    state.problemCodeEditor?.focus();

    const languageSelect = document.getElementById("problem-language-select");
    let currentLanguage = selectedLanguage;
    languageSelect?.addEventListener("change", (event) => {
      const nextLanguage = event.currentTarget.value;
      const previousTemplate = getDefaultCodeTemplate(currentLanguage);
      if (!codeEditor.value.trim() || codeEditor.value === previousTemplate) {
        const nextTemplate = getDefaultCodeTemplate(nextLanguage);
        if (state.problemCodeEditor) {
          state.problemCodeEditor.setValue(nextTemplate);
        } else {
          codeEditor.value = nextTemplate;
        }
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
      state.runResultPending = true;
      refreshRunResultPanel();

      try {
        const result = await apiFetch("/submissions/run", {
          method: "POST",
          body: JSON.stringify({
            problem_id: Number(problem.id),
            language,
            code,
          }),
        });
        state.runResultPending = false;
        state.runResult = { problemID: problem.id, contestID: null, ...result };
        refreshRunResultPanel();
      } catch (err) {
        state.runResultPending = false;
        refreshRunResultPanel();
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

  if (state.runResultUIAbort) {
    state.runResultUIAbort.abort();
  }
  const controller = new AbortController();
  const { signal } = controller;
  state.runResultUIAbort = controller;

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
  }, { signal });

  window.addEventListener("mousemove", onMove, { signal });
  window.addEventListener("mouseup", stopDragging, { signal });
  window.addEventListener("mouseleave", stopDragging, { signal });
}

function renderProblemBlock(title, content) {
  return `
    <div class="detail-block">
      <h3>${escapeHTML(title)}</h3>
      ${renderMarkdownBlock(content)}
    </div>
  `;
}

function renderProblemCodeBlock(title, content) {
  return `
    <div class="detail-block">
      <h3>${escapeHTML(title)}</h3>
      ${renderPlainCodeBlock(content)}
    </div>
  `;
}
function renderProblemSolutions(solutions, problemID) {
  const list = solutions || [];
  return `
    <div class="detail-block">
      <div class="view-header compact">
        <div>
          <h3>Solutions</h3>
          <p class="view-subtitle">Editorial notes and official write-ups for this problem.</p>
        </div>
        ${isTeacherUser() ? `<a class="ghost-button" href="#/teacher/problems/${problemID}/solutions">Manage Solutions</a>` : ""}
      </div>
      ${list.length ? `
        <div class="solution-stack">
          ${list.map((item) => `
            <article class="solution-card">
              <div class="view-header compact">
                <div>
                  <h4 class="solution-title">${escapeHTML(item.title)}</h4>
                  <p class="view-subtitle">Solution #${item.id}</p>
                </div>
                <span class="status-pill ${teachingVisibilityClass(item.visibility)}">${escapeHTML(item.visibility)}</span>
              </div>
              ${renderMarkdownBlock(item.content || "")}
            </article>
          `).join("")}
        </div>
      ` : renderTeachingEmpty("No public solutions yet.")}
    </div>
  `;
}

function renderPlainCodeBlock(content) {
  return `<pre class="problem-plain-pre">${escapeHTML(content || "")}</pre>`;
}

function renderMarkdownBlock(content) {
  const source = (content || "").toString().trim();
  if (!source) {
    return `<div class="problem-markdown empty">No content.</div>`;
  }
  return `<div class="problem-markdown">${renderMarkdown(source)}</div>`;
}

function renderMarkdown(content) {
  const lines = content.replace(/\r\n?/g, "\n").split("\n");
  const html = [];
  let paragraph = [];
  let listType = null;
  let listItems = [];
  let codeFence = null;
  let codeLines = [];

  const flushParagraph = () => {
    if (!paragraph.length) return;
    html.push(`<p>${renderInlineMarkdown(paragraph.join("<br>"))}</p>`);
    paragraph = [];
  };

  const flushList = () => {
    if (!listType || !listItems.length) return;
    html.push(`<${listType}>${listItems.map((item) => `<li>${renderInlineMarkdown(item)}</li>`).join("")}</${listType}>`);
    listType = null;
    listItems = [];
  };

  const flushCodeFence = () => {
    if (!codeFence) return;
    html.push(`<pre class="problem-markdown-pre"><code>${escapeHTML(codeLines.join("\n"))}</code></pre>`);
    codeFence = null;
    codeLines = [];
  };

  for (const rawLine of lines) {
    const line = rawLine ?? "";

    if (codeFence) {
      if (/^```/.test(line.trim())) {
        flushCodeFence();
      } else {
        codeLines.push(line);
      }
      continue;
    }

    if (/^```/.test(line.trim())) {
      flushParagraph();
      flushList();
      codeFence = "fence";
      codeLines = [];
      continue;
    }

    if (!line.trim()) {
      flushParagraph();
      flushList();
      continue;
    }

    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      flushParagraph();
      flushList();
      const level = headingMatch[1].length;
      html.push(`<h${level + 1}>${renderInlineMarkdown(headingMatch[2].trim())}</h${level + 1}>`);
      continue;
    }

    const bulletMatch = line.match(/^\s*[-*]\s+(.*)$/);
    if (bulletMatch) {
      flushParagraph();
      if (listType && listType !== "ul") {
        flushList();
      }
      listType = "ul";
      listItems.push(bulletMatch[1].trim());
      continue;
    }

    const orderedMatch = line.match(/^\s*\d+\.\s+(.*)$/);
    if (orderedMatch) {
      flushParagraph();
      if (listType && listType !== "ol") {
        flushList();
      }
      listType = "ol";
      listItems.push(orderedMatch[1].trim());
      continue;
    }

    const quoteMatch = line.match(/^\s*>\s?(.*)$/);
    if (quoteMatch) {
      flushParagraph();
      flushList();
      html.push(`<blockquote>${renderInlineMarkdown(quoteMatch[1])}</blockquote>`);
      continue;
    }

    paragraph.push(line);
  }

  flushParagraph();
  flushList();
  flushCodeFence();

  return html.join("");
}

function renderInlineMarkdown(input) {
  const escaped = escapeHTML(input);
  return escaped
    .replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noreferrer">$1</a>')
    .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
    .replace(/(^|[^\*])\*([^*]+)\*/g, "$1<em>$2</em>")
    .replace(/`([^`]+)`/g, "<code>$1</code>");
}
function renderRunResultPanel() {
  const result = state.runResult;
  const statusText = state.runResultPending
    ? "Running"
    : (result?.status || "Not Run");
  const body = state.runResultPending
    ? `<div class="run-result-empty">Running your code against the sample tests...</div>`
    : !result
      ? `<div class="run-result-empty">Click Run to execute your code against the sample tests.</div>`
      : `
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
        `).join("") || `<div class="run-result-empty">Run finished, but no sample case details were returned.</div>`}
      `;

  return `
    <div class="run-result-panel" id="run-result-panel" style="height:${state.runResultHeight}px;">
      <div class="run-result-resizer" id="run-result-resizer" aria-hidden="true"></div>
      <div class="run-result-head">
        <strong>Run Result</strong>
        <span class="status-pill ${state.runResultPending ? "status-pending" : statusClass(statusText)}">${escapeHTML(statusText)}</span>
      </div>
      ${body}
    </div>
  `;
}

function refreshRunResultPanel() {
  const slot = document.getElementById("run-result-slot");
  const runButton = document.getElementById("run-sample-btn");
  if (runButton) {
    runButton.disabled = state.runResultPending;
    runButton.textContent = state.runResultPending ? "Running..." : "Run";
  }
  if (!slot) {
    return;
  }
  slot.innerHTML = renderRunResultPanel();
  initRunResultUI();
}

async function mountProblemCodeEditor(textarea) {
  if (!textarea || state.problemCodeEditor) {
    return state.problemCodeEditor;
  }

  try {
    await (window.codeMirrorReadyPromise || Promise.resolve());
  } catch (error) {
    console.error("[problem-editor-init]", error);
    setFlash(`CodeMirror load failed: ${error?.message || error}`, true);
    return null;
  }

  if (!document.body.contains(textarea) || typeof window.createProblemCodeEditor !== "function") {
    return null;
  }

  state.problemCodeEditor = window.createProblemCodeEditor(textarea);
  return state.problemCodeEditor;
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

async function loadRecentProblemSubmissions(problemID, contestID = null) {
  try {
    const query = new URLSearchParams({ page: "1", page_size: "5", problem_id: String(problemID) });
    if (contestID) {
      query.set("contest_id", String(contestID));
    }
    const data = await apiFetch(`/submissions/my?${query.toString()}`, { method: "GET" });
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

function renderProblemRecentSubmissions(problemID, submissions, contestID = null) {
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
        <a class="ghost-button" href="#/submissions?problem_id=${problemID}${contestID ? `&contest_id=${contestID}` : ``}">View All</a>
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


