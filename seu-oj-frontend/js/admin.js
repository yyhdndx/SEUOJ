// Admin domain pages and helpers
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
          <a class="ghost-button" href="#/teacher/problems/${problem.id}/solutions">Solutions</a>
          <a class="primary-button" href="#/admin/problems/${problem.id}/edit">Edit</a>
        </div>
      </div>
      <section class="detail-grid">
        <article class="detail-card">
          ${renderProblemBlock("Description", problem.description)}
          ${renderProblemBlock("Input", problem.input_desc)}
          ${renderProblemBlock("Output", problem.output_desc)}
          ${renderProblemCodeBlock("Sample Input", problem.sample_input)}
          ${renderProblemCodeBlock("Sample Output", problem.sample_output)}
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


async function renderAdminSubmissions() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before managing submissions.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access submission management.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading admin submissions...</p></div>`;
  try {
    const queryString = getCurrentHashPath().split("?")[1] || "";
    const params = new URLSearchParams(queryString);
    const page = params.get("page") || "1";
    const pageSize = params.get("page_size") || "20";
    const userID = params.get("user_id") || "";
    const problemID = params.get("problem_id") || "";
    const contestID = params.get("contest_id") || "";
    const status = params.get("status") || "";

    const query = new URLSearchParams({ page, page_size: pageSize });
    if (userID) query.set("user_id", userID);
    if (problemID) query.set("problem_id", problemID);
    if (contestID) query.set("contest_id", contestID);
    if (status) query.set("status", status);

    const [adminStats, data] = await Promise.all([
      apiFetch("/stats/admin", { method: "GET" }),
      apiFetch(`/admin/submissions?${query.toString()}`, { method: "GET" }),
    ]);

    const list = data.list || [];
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Manage Submissions</h1>
          <p class="view-subtitle">Observe queue pressure, filter submissions, inspect contest-tagged runs, and trigger rejudge when needed.</p>
        </div>
      </div>
      <section class="detail-card">
        <div class="verdict-summary-grid">
          <div class="verdict-summary-card"><span class="status-pill status-pending">Queue</span><strong>${adminStats.queue_length}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-neutral">Submissions</span><strong>${adminStats.submissions_total}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-pending">Pending / Running</span><strong>${adminStats.pending_submissions + adminStats.running_submissions}</strong></div>
          <div class="verdict-summary-card"><span class="status-pill status-error">System Errors</span><strong>${adminStats.system_errors}</strong></div>
        </div>
      </section>
      <form id="admin-submission-filters" class="toolbar" style="margin-top:18px;">
        <div>
          <label class="field-label">User ID</label>
          <input class="text-input" name="user_id" value="${escapeHTML(userID)}" />
        </div>
        <div>
          <label class="field-label">Problem ID</label>
          <input class="text-input" name="problem_id" value="${escapeHTML(problemID)}" />
        </div>
        <div>
          <label class="field-label">Contest ID</label>
          <input class="text-input" name="contest_id" value="${escapeHTML(contestID)}" placeholder="optional" />
        </div>
        <div>
          <label class="field-label">Status</label>
          <input class="text-input" name="status" value="${escapeHTML(status)}" placeholder="Accepted / Pending ..." />
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
          <button class="ghost-button" type="button" id="clear-admin-submission-filters">Clear</button>
        </div>
      </form>
      <section class="detail-card" style="margin-top:18px;">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Problem</th>
              <th>Contest</th>
              <th>Lang</th>
              <th>Status</th>
              <th>Passed</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            ${list.map((item) => `
              <tr>
                <td>${item.id}</td>
                <td>${item.user_id}</td>
                <td>${item.problem_id}</td>
                <td>${item.contest_id ? `<div style="display:flex; gap:6px; align-items:center; flex-wrap:wrap;"><a class="table-link" href="#/contests/${item.contest_id}">${item.contest_id}</a>${renderContestModeBadge(item)}</div>` : "-"}</td>
                <td>${escapeHTML(item.language)}</td>
                <td><span class="status-pill ${statusClass(item.status)}">${escapeHTML(item.status)}</span></td>
                <td>${item.passed_count}/${item.total_count}</td>
                <td class="mono">${escapeHTML(item.created_at)}</td>
                <td>
                  <a class="table-link" href="#/submissions/${item.id}">View</a>
                  <span class="mono"> | </span>
                  <button class="link-button admin-rejudge-btn" data-submission-id="${item.id}">Rejudge</button>
                </td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      </section>
      <section class="detail-grid" style="margin-top:18px;">
        <article class="detail-card">
          <h3>Status Breakdown</h3>
          ${renderCountTable(adminStats.status_breakdown || [], "Status", "Count")}
        </article>
        <article class="detail-card">
          <h3>Inventory</h3>
          <div class="metric-list">
            <div class="metric"><span class="metric-label">Users</span><span class="metric-value">${adminStats.users_total}</span></div>
            <div class="metric"><span class="metric-label">Problems</span><span class="metric-value">${adminStats.problems_total}</span></div>
            <div class="metric"><span class="metric-label">Hidden Problems</span><span class="metric-value">${adminStats.hidden_problems}</span></div>
          </div>
        </article>
      </section>
    `;

    document.getElementById("admin-submission-filters").addEventListener("submit", (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const next = new URLSearchParams();
      ["user_id", "problem_id", "contest_id", "status", "page", "page_size"].forEach((key) => {
        const value = (form.get(key) || "").toString().trim();
        if (value) next.set(key, value);
      });
      location.hash = next.toString() ? `#/admin/submissions?${next.toString()}` : "#/admin/submissions";
    });

    document.getElementById("clear-admin-submission-filters").addEventListener("click", () => {
      location.hash = "#/admin/submissions";
    });

    document.querySelectorAll(".admin-rejudge-btn").forEach((button) => {
      button.addEventListener("click", async () => {
        const id = button.dataset.submissionId;
        try {
          await apiFetch(`/admin/submissions/${id}/rejudge`, { method: "POST" });
          setFlash(`Submission #${id} requeued`, false);
          await renderAdminSubmissions();
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load admin submissions failed: ${escapeHTML(err.message)}</p></div>`;
  }
}


async function renderAdminUsers() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before managing users.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== 'admin') {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access user management.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading users...</p></div>`;
  try {
    const data = await apiFetch('/admin/users?page=1&page_size=100', { method: 'GET' });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Manage Users</h1>
          <p class="view-subtitle">Adjust roles and account status directly.</p>
        </div>
      </div>
      <section class="detail-card">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Username</th>
              <th>Student ID</th>
              <th>Role</th>
              <th>Status</th>
              <th>Save</th>
            </tr>
          </thead>
          <tbody>
            ${(data.list || []).map((item) => `
              <tr>
                <td>${item.id}</td>
                <td><input class="text-input admin-user-input" data-field="username" data-user-id="${item.id}" value="${escapeHTML(item.username)}" /></td>
                <td><input class="text-input admin-user-input" data-field="userid" data-user-id="${item.id}" value="${escapeHTML(item.userid)}" /></td>
                <td>
                  <select class="select-input admin-user-input" data-field="role" data-user-id="${item.id}">
                    ${['student', 'admin', 'teacher'].map((role) => `<option value="${role}" ${item.role === role ? 'selected' : ''}>${role}</option>`).join('')}
                  </select>
                </td>
                <td>
                  <select class="select-input admin-user-input" data-field="status" data-user-id="${item.id}">
                    ${['active', 'disabled'].map((status) => `<option value="${status}" ${item.status === status ? 'selected' : ''}>${status}</option>`).join('')}
                  </select>
                </td>
                <td><button class="ghost-button admin-user-save" data-user-id="${item.id}">Save</button></td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </section>
    `;

    document.querySelectorAll('.admin-user-save').forEach((button) => {
      button.addEventListener('click', async () => {
        const id = button.dataset.userId;
        const read = (field) => document.querySelector(`.admin-user-input[data-user-id="${id}"][data-field="${field}"]`).value;
        try {
          await apiFetch(`/admin/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify({
              username: read('username'),
              userid: read('userid'),
              role: read('role'),
              status: read('status'),
            }),
          });
          setFlash(`User #${id} updated`, false);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load users failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAdminAnnouncements() {
  if (!state.token) {
    app.innerHTML = `<div class="detail-card"><p>Please login before managing announcements.</p></div>`;
    return;
  }
  if (!state.user || state.user.role !== 'admin') {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access announcement management.</p></div>`;
    return;
  }

  app.innerHTML = `<div class="detail-card"><p>Loading announcements...</p></div>`;
  try {
    const data = await apiFetch('/announcements?page=1&page_size=50', { method: 'GET' });
    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Manage Announcements</h1>
          <p class="view-subtitle">Publish updates that appear on the public announcement feed.</p>
        </div>
        <div>
          <a class="primary-button" href="#/admin/announcements/new">New Announcement</a>
        </div>
      </div>
      <section class="detail-card">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Title</th>
              <th>Pinned</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            ${(data.list || []).map((item) => `
              <tr>
                <td>${item.id}</td>
                <td>${escapeHTML(item.title)}</td>
                <td>${item.is_pinned ? 'Yes' : 'No'}</td>
                <td class="mono">${escapeHTML(item.created_at)}</td>
                <td>
                  <a class="table-link" href="#/admin/announcements/${item.id}/edit">Edit</a>
                  <span class="mono"> | </span>
                  <button class="link-button admin-announcement-delete" data-announcement-id="${item.id}">Delete</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </section>
    `;

    document.querySelectorAll('.admin-announcement-delete').forEach((button) => {
      button.addEventListener('click', async () => {
        const id = button.dataset.announcementId;
        try {
          await apiFetch(`/admin/announcements/${id}`, { method: 'DELETE' });
          setFlash(`Announcement #${id} deleted`, false);
          await renderAdminAnnouncements();
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load admin announcements failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAdminAnnouncementCreate() {
  return renderAdminAnnouncementForm(null, null);
}

async function renderAdminAnnouncementEdit(id) {
  let initial = null;
  try {
    initial = await apiFetch(`/announcements/${id}`, { method: 'GET' });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load announcement failed: ${escapeHTML(err.message)}</p></div>`;
    return;
  }
  return renderAdminAnnouncementForm(id, initial);
}

async function renderAdminAnnouncementForm(id, initial) {
  if (!state.token || state.user?.role !== 'admin') {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access announcement editing.</p></div>`;
    return;
  }
  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">${id ? 'Edit Announcement' : 'New Announcement'}</h1>
        <p class="view-subtitle">Short operational updates, release notes, or maintenance notices.</p>
      </div>
    </div>
    <form id="announcement-form" class="detail-card toolbar">
      <div class="full">
        <label class="field-label">Title</label>
        <input class="text-input" name="title" value="${escapeHTML(initial?.title || '')}" required />
      </div>
      <div class="full">
        <label class="field-label">Content</label>
        <textarea class="text-area" name="content" required>${escapeHTML(initial?.content || '')}</textarea>
      </div>
      <div>
        <label class="field-label">Pinned</label>
        <input type="checkbox" name="is_pinned" ${initial?.is_pinned ? 'checked' : ''} />
      </div>
      <div class="full" style="display:flex; gap:10px; align-items:center;">
        <button class="primary-button" type="submit">${id ? 'Update' : 'Create'}</button>
        <a class="ghost-button" href="#/admin/announcements">Back</a>
      </div>
    </form>
  `;

  document.getElementById('announcement-form').addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await apiFetch(id ? `/admin/announcements/${id}` : '/admin/announcements', {
        method: id ? 'PUT' : 'POST',
        body: JSON.stringify({
          title: form.get('title'),
          content: form.get('content'),
          is_pinned: form.get('is_pinned') === 'on',
        }),
      });
      setFlash(`Announcement ${id ? 'updated' : 'created'}`, false);
      location.hash = '#/admin/announcements';
    } catch (err) {
      setFlash(err.message, true);
    }
  });
}




async function renderAdminContestAnnouncementCreate(contestID) {
  return renderAdminContestAnnouncementForm(contestID, null, null);
}

async function renderAdminContestAnnouncementEdit(contestID, announcementID) {
  if (!state.token || state.user?.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can edit contest announcements.</p></div>`;
    return;
  }
  try {
    const initial = await apiFetch(`/admin/contests/${contestID}/announcements/${announcementID}`, { method: "GET" });
    return renderAdminContestAnnouncementForm(contestID, announcementID, initial);
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load contest announcement failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderAdminContestAnnouncementForm(contestID, announcementID, initial) {
  if (!state.token || state.user?.role !== "admin") {
    app.innerHTML = `<div class="detail-card"><p>Only admin can access contest announcement editing.</p></div>`;
    return;
  }
  app.innerHTML = `
    <div class="view-header">
      <div>
        <h1 class="view-title">${announcementID ? "Edit Contest Announcement" : "New Contest Announcement"}</h1>
        <p class="view-subtitle">Contest-scoped notices for clarifications, data fixes, and schedule updates.</p>
      </div>
    </div>
    <form id="contest-announcement-form" class="detail-card toolbar">
      <div class="full">
        <label class="field-label">Title</label>
        <input class="text-input" name="title" value="${escapeHTML(initial?.title || "")}" required />
      </div>
      <div class="full">
        <label class="field-label">Content</label>
        <textarea class="text-area" name="content" required>${escapeHTML(initial?.content || "")}</textarea>
      </div>
      <div>
        <label class="field-label">Pinned</label>
        <input type="checkbox" name="is_pinned" ${initial?.is_pinned ? "checked" : ""} />
      </div>
      <div class="full" style="display:flex; gap:10px; align-items:center;">
        <button class="primary-button" type="submit">${announcementID ? "Update" : "Create"}</button>
        <a class="ghost-button" href="#/admin/contests/${contestID}">Back</a>
      </div>
    </form>
  `;

  document.getElementById("contest-announcement-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await apiFetch(announcementID ? `/admin/contests/${contestID}/announcements/${announcementID}` : `/admin/contests/${contestID}/announcements`, {
        method: announcementID ? "PUT" : "POST",
        body: JSON.stringify({
          title: form.get("title"),
          content: form.get("content"),
          is_pinned: form.get("is_pinned") === "on",
        }),
      });
      setFlash(`Contest announcement ${announcementID ? "updated" : "created"}`, false);
      location.hash = `#/admin/contests/${contestID}`;
    } catch (err) {
      setFlash(err.message, true);
    }
  });
}


