function solutionVisibilityClass(visibility) {
  if (visibility === "public") return "status-accepted";
  if (visibility === "class") return "status-pending";
  return "status-neutral";
}

function renderSolutionEmpty(message) {
  return `<div class="solution-empty">${escapeHTML(message)}</div>`;
}

function renderSolutionPreview(content) {
  if (typeof renderMarkdownBlock === "function") {
    return renderMarkdownBlock(content);
  }
  return `<div class="problem-markdown">${escapeHTML(content || "No content.")}</div>`;
}

function solutionDraftForUser(list) {
  return list.find((item) => Number(item.author_id) === Number(state.user?.id)) || null;
}

function buildSolutionManagerHash(problemID, scope, solutionID = "") {
  const query = new URLSearchParams();
  if (scope) {
    query.set("scope", scope);
  }
  if (solutionID) {
    query.set("solution", String(solutionID));
  }
  const queryString = query.toString();
  return `#/problems/${problemID}/solutions/manage${queryString ? `?${queryString}` : ""}`;
}

function nextSolutionIDAfterDelete(list, deletedSolutionID, fallbackID = "") {
  const remaining = list.filter((item) => String(item.id) !== String(deletedSolutionID));
  if (!remaining.length) return "";
  const deletedIndex = list.findIndex((item) => String(item.id) === String(deletedSolutionID));
  return remaining[deletedIndex]?.id || remaining[deletedIndex - 1]?.id || remaining[0]?.id || fallbackID || "";
}

function solutionNavigationFor(list, selectedSolutionID) {
  const index = list.findIndex((item) => String(item.id) === String(selectedSolutionID));
  if (index < 0) {
    return { index: -1, total: list.length, previousID: "", nextID: "" };
  }
  return {
    index,
    total: list.length,
    previousID: list[index - 1]?.id || "",
    nextID: list[index + 1]?.id || "",
  };
}

function renderSolutionForm(problemID, solution) {
  const isEditing = !!solution;
  return `
    <form id="problem-solution-form" class="solution-form" data-solution-id="${isEditing ? solution.id : ""}" data-author-id="${isEditing ? solution.author_id : ""}">
      <input type="hidden" name="solution_id" value="${isEditing ? solution.id : ""}" />
      <label class="field-label">Title</label>
      <input class="text-input" name="title" value="${escapeHTML(solution?.title || "")}" required />
      <label class="field-label">Visibility</label>
      <select class="select-input" name="visibility">
        <option value="public" ${solution?.visibility === "public" ? "selected" : ""}>public</option>
        <option value="private" ${solution?.visibility === "private" ? "selected" : ""}>private</option>
      </select>
      <label class="field-label">Content</label>
      <textarea class="text-area solution-editor-source" name="content" rows="20" required>${escapeHTML(solution?.content || "")}</textarea>
      <div class="view-subtitle">Markdown is supported. Users need an Accepted submission on this problem before publishing.</div>
      <div class="solution-form-actions">
        <button class="primary-button" type="submit">${isEditing ? "Save Solution" : "Create Solution"}</button>
        ${isEditing ? `<button class="ghost-button solution-delete-current" type="button" data-solution-id="${solution.id}">Delete</button>` : ""}
        <a class="ghost-button" href="#/problems/${problemID}">Cancel</a>
      </div>
    </form>
  `;
}

function renderAdminSolutionList(list, selectedSolutionID = "") {
  if (!state.user || state.user.role !== "admin") return "";
  return `
    <div class="solution-admin-list">
      <div class="solution-admin-list-head">
        <h4>All Solutions</h4>
        <span class="status-pill status-neutral">${list.length}</span>
      </div>
      ${list.length ? `
        <div class="solution-admin-items">
          ${list.map((item) => `
            <article class="solution-admin-item ${Number(selectedSolutionID) === Number(item.id) ? "is-active" : ""}">
              <div>
                <strong>${escapeHTML(item.title)}</strong>
                <p class="view-subtitle">updated ${escapeHTML(item.updated_at)}</p>
              </div>
              <div class="solution-admin-item-actions">
                <span class="status-pill ${solutionVisibilityClass(item.visibility)}">${escapeHTML(item.visibility)}</span>
                <button class="ghost-button solution-load-button" type="button"
                  data-solution-id="${item.id}"
                  data-author-id="${item.author_id}"
                  data-title="${encodeURIComponent(item.title || "")}"
                  data-visibility="${escapeHTML(item.visibility || "public")}"
                  data-content="${encodeURIComponent(item.content || "")}">${Number(selectedSolutionID) === Number(item.id) ? "Current" : "Edit"}</button>
                <button class="ghost-button solution-delete-button" type="button" data-solution-id="${item.id}">Delete</button>
              </div>
            </article>
          `).join("")}
        </div>
      ` : renderSolutionEmpty("No solutions yet.")}
    </div>
  `;
}

async function renderProblemSolutionManager(problemID, scope = "my") {
  if (!state.user) {
    app.innerHTML = `<div class="detail-card"><p>Login required.</p></div>`;
    return;
  }
  app.innerHTML = `<div class="detail-card"><p>Loading problem solutions...</p></div>`;
  try {
    const problemPath = state.user?.role === "admin" ? `/admin/problems/${problemID}` : `/problems/${problemID}`;
    const [problem, solutions] = await Promise.all([
      apiFetch(problemPath, { method: "GET" }),
      apiFetch(`/problems/${problemID}/solutions/manage`, { method: "GET" }),
    ]);
    const list = solutions.list || [];
    const canManageAll = state.user?.role === "admin";
    const queryParams = new URLSearchParams(getCurrentHashPath().split("?")[1] || "");
    const activeScope = canManageAll && queryParams.get("scope") === "all" ? "all" : "my";
    const ownSolution = solutionDraftForUser(list);
    const selectedSolutionID = queryParams.get("solution") || "";
    const selectedSolution = activeScope === "all"
      ? (list.find((item) => String(item.id) === String(selectedSolutionID)) || list[0] || null)
      : (ownSolution || null);
    const activeSolutionID = selectedSolution?.id || "";
    const solutionNav = activeScope === "all" ? solutionNavigationFor(list, activeSolutionID) : null;
    app.innerHTML = `
      <div class="solution-manager-page">
        <header class="solution-manager-bar">
          <div>
            <h1>Solution Manager</h1>
            <p>${escapeHTML(problem.title)}</p>
          </div>
          <div class="solution-manager-actions">
            ${canManageAll ? `
              <button class="ghost-button solution-scope-button ${activeScope === "my" ? "is-active" : ""}" type="button" data-solution-scope="my" data-solution-id="${ownSolution?.id || activeSolutionID || ""}">My Solution</button>
              <button class="ghost-button solution-scope-button ${activeScope === "all" ? "is-active" : ""}" type="button" data-solution-scope="all" data-solution-id="${activeSolutionID || ownSolution?.id || list[0]?.id || ""}">Manage Solutions</button>
            ` : ""}
            <a class="ghost-button" href="#/problems/${problem.id}">Back to Problem</a>
          </div>
        </header>
        <section class="solution-manager-layout">
          <section class="detail-card solution-manager-section solution-editor-section">
            <div class="solution-section-heading">
              <h3>${selectedSolution ? "Edit" : "Create"}</h3>
              ${selectedSolution ? '<span class="status-pill status-neutral">Current Solution</span>' : ""}
            </div>
            ${solutionNav && solutionNav.total > 1 ? `
              <div class="solution-editor-nav">
                <span class="view-subtitle">${solutionNav.index + 1} / ${solutionNav.total}</span>
                <div class="solution-editor-nav-actions">
                  <button class="ghost-button solution-nav-button" type="button" data-solution-target="${solutionNav.previousID}" ${solutionNav.previousID ? "" : "disabled"}>Previous</button>
                  <button class="ghost-button solution-nav-button" type="button" data-solution-target="${solutionNav.nextID}" ${solutionNav.nextID ? "" : "disabled"}>Next</button>
                </div>
              </div>
            ` : ""}
            ${renderSolutionForm(problem.id, selectedSolution)}
            ${activeScope === "all" ? renderAdminSolutionList(list, activeSolutionID) : ""}
          </section>
          <section class="detail-card solution-manager-section solution-preview-section">
            <div class="solution-section-heading">
              <h3>${activeScope === "all" ? "Selected Solution" : "My Solution"}</h3>
              <span class="view-subtitle">Preview</span>
            </div>
            <div class="solution-preview-title">${escapeHTML(selectedSolution?.title || "Untitled Solution")}</div>
            <div class="solution-preview-meta">
              <span class="status-pill ${solutionVisibilityClass(selectedSolution?.visibility || "public")}">${escapeHTML(selectedSolution?.visibility || "public")}</span>
              <span>${selectedSolution ? "" : "Draft"}</span>
            </div>
            <div class="solution-preview-body">${renderSolutionPreview(selectedSolution?.content || "")}</div>
          </section>
        </section>
      </div>
    `;

    setupSolutionEditor(problemID, activeScope, activeSolutionID, list);
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load problem solutions failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

function setupSolutionEditor(problemID, scope, selectedSolutionID, list) {
  const form = document.getElementById("problem-solution-form");
  const titleInput = form?.querySelector('[name="title"]');
  const visibilityInput = form?.querySelector('[name="visibility"]');
  const contentInput = form?.querySelector('[name="content"]');
  const previewTitle = document.querySelector(".solution-preview-title");
  const previewMeta = document.querySelector(".solution-preview-meta");
  const previewBody = document.querySelector(".solution-preview-body");
  let markdownEditor = null;

  const refreshPreview = () => {
    const title = titleInput?.value.trim() || "Untitled Solution";
    const visibility = visibilityInput?.value || "public";
    const authorID = form?.dataset.authorId || "";
    if (previewTitle) previewTitle.textContent = title;
    if (previewMeta) {
      previewMeta.innerHTML = `
        <span class="status-pill ${solutionVisibilityClass(visibility)}">${escapeHTML(visibility)}</span>
        <span>${authorID ? "Author" : "Draft"}</span>
      `;
    }
    if (previewBody && contentInput) {
      previewBody.innerHTML = renderSolutionPreview(contentInput.value);
    }
  };

  [titleInput, visibilityInput, contentInput].forEach((input) => {
    input?.addEventListener("input", refreshPreview);
    input?.addEventListener("change", refreshPreview);
  });

  const attachMarkdownEditor = () => {
    if (!contentInput || typeof window.createSolutionMarkdownEditor !== "function") return;
    markdownEditor = window.createSolutionMarkdownEditor(contentInput, {
      onChange: refreshPreview,
    });
  };

  if (window.codeMirrorReadyPromise) {
    window.codeMirrorReadyPromise.then(attachMarkdownEditor).catch(() => {});
  } else {
    attachMarkdownEditor();
  }

  form?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const formData = new FormData(form);
    const solutionID = String(form.dataset.solutionId || "").trim();
    const method = solutionID ? "PUT" : "POST";
    const path = solutionID ? `/problems/${problemID}/solutions/${solutionID}` : `/problems/${problemID}/solutions`;
    try {
      const result = await apiFetch(path, {
        method,
        body: JSON.stringify({
          title: formData.get("title"),
          visibility: formData.get("visibility"),
          content: formData.get("content"),
        }),
      });
      setFlash(solutionID ? "Solution updated" : "Solution created", false);
      const nextSolutionID = String(result?.id || solutionID || selectedSolutionID || "");
      const nextHash = buildSolutionManagerHash(problemID, scope === "all" ? "all" : "my", nextSolutionID);
      if (location.hash === nextHash) {
        return renderProblemSolutionManager(problemID, scope);
      }
      location.hash = nextHash;
    } catch (err) {
      setFlash(err.message, true);
    }
  });

  document.querySelectorAll(".solution-load-button").forEach((button) => {
    button.addEventListener("click", () => {
      const solutionID = button.dataset.solutionId || "";
      location.hash = buildSolutionManagerHash(problemID, scope === "all" ? "all" : "my", solutionID);
    });
  });

  document.querySelectorAll(".solution-scope-button").forEach((button) => {
    button.addEventListener("click", () => {
      const nextScope = button.dataset.solutionScope || "my";
      const nextSolutionID = button.dataset.solutionId || "";
      location.hash = buildSolutionManagerHash(problemID, nextScope, nextSolutionID);
    });
  });

  document.querySelectorAll(".solution-nav-button").forEach((button) => {
    button.addEventListener("click", () => {
      const targetSolutionID = button.dataset.solutionTarget || "";
      if (!targetSolutionID) return;
      location.hash = buildSolutionManagerHash(problemID, "all", targetSolutionID);
    });
  });

  document.querySelectorAll(".solution-delete-button, .solution-delete-current").forEach((button) => {
    button.addEventListener("click", async () => {
      const solutionID = button.dataset.solutionId;
      const confirmed = window.confirm("Delete this solution?");
      if (!confirmed) return;
      try {
        await apiFetch(`/problems/${problemID}/solutions/${solutionID}`, { method: "DELETE" });
        setFlash("Solution deleted", false);
        const nextSolutionID = scope === "all"
          ? nextSolutionIDAfterDelete(list, solutionID, selectedSolutionID)
          : "";
        const nextHash = buildSolutionManagerHash(problemID, scope === "all" ? "all" : "my", nextSolutionID);
        if (location.hash === nextHash) {
          return renderProblemSolutionManager(problemID, scope);
        }
        location.hash = nextHash;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  });
}
