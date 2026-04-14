function forumScopeLabel(scopeType, scopeID) {
  if (scopeType === "problem") return scopeID ? `Problem #${scopeID}` : "Problem";
  if (scopeType === "contest") return scopeID ? `Contest #${scopeID}` : "Contest";
  return "General";
}

function forumScopeLink(scopeType, scopeID) {
  if (!scopeID) return "#/forum";
  if (scopeType === "problem") return `#/problems/${scopeID}`;
  if (scopeType === "contest") return `#/contests/${scopeID}`;
  return "#/forum";
}

function canManageForumContent(authorID) {
  return !!state.user && (Number(state.user.id) === Number(authorID) || state.user.role === "admin" || state.user.role === "teacher");
}

function canModerateForum() {
  return !!state.user && (state.user.role === "admin" || state.user.role === "teacher");
}

function forumStatusPills(topic) {
  const pills = [`<span class="status-pill status-neutral">${escapeHTML(forumScopeLabel(topic.scope_type, topic.scope_id))}</span>`];
  if (topic.is_pinned) pills.push(`<span class="status-pill status-accepted">Pinned</span>`);
  if (topic.is_locked) pills.push(`<span class="status-pill status-pending">Locked</span>`);
  pills.push(`<span class="status-pill status-neutral">${topic.reply_count} replies</span>`);
  return pills.join("");
}

function renderForumSummaryCards(list, scopeType) {
  const total = list.length;
  const pinned = list.filter((item) => item.is_pinned).length;
  const locked = list.filter((item) => item.is_locked).length;
  const active = list.filter((item) => item.reply_count > 0).length;
  return `
    <div class="forum-summary-grid">
      <article class="forum-summary-card">
        <span class="forum-summary-label">Scope</span>
        <strong class="forum-summary-value">${escapeHTML(scopeType ? forumScopeLabel(scopeType) : "All Topics")}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Topics</span>
        <strong class="forum-summary-value">${total}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Pinned</span>
        <strong class="forum-summary-value">${pinned}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Active</span>
        <strong class="forum-summary-value">${active}</strong>
      </article>
      <article class="forum-summary-card">
        <span class="forum-summary-label">Locked</span>
        <strong class="forum-summary-value">${locked}</strong>
      </article>
    </div>
  `;
}

function renderForumTopicCards(list) {
  if (!list.length) {
    return `<div class="detail-card"><p>No topics yet.</p></div>`;
  }
  return `
    <div class="forum-topic-list">
      ${list.map((item) => `
        <article class="forum-topic-card ${item.is_pinned ? "pinned" : ""} ${item.is_locked ? "locked" : ""}">
          <div class="view-header compact">
            <div>
              <h3 class="forum-topic-title"><a class="table-link" href="#/forum/topics/${item.id}">${escapeHTML(item.title)}</a></h3>
              <p class="view-subtitle">${escapeHTML(item.author_name)} · ${escapeHTML(item.created_at)}${item.last_reply_at ? ` · last reply ${escapeHTML(item.last_reply_at)}` : ""}</p>
            </div>
            <div class="pill-row">${forumStatusPills(item)}</div>
          </div>
          <p class="forum-topic-preview">${escapeHTML(item.content_preview || "")}</p>
          <div class="forum-topic-footer">
            <a class="ghost-button" href="${forumScopeLink(item.scope_type, item.scope_id)}">Open Scope</a>
            <a class="ghost-button" href="#/forum/topics/${item.id}">View Topic</a>
          </div>
        </article>
      `).join("")}
    </div>
  `;
}

async function renderForum() {
  const query = getHashQueryParams();
  const page = Number(query.get("page") || 1);
  const pageSize = Number(query.get("page_size") || 20);
  const keyword = (query.get("keyword") || "").trim();
  const scopeType = (query.get("scope_type") || "").trim();
  const scopeID = (query.get("scope_id") || "").trim();

  app.innerHTML = `<div class="detail-card"><p>Loading forum...</p></div>`;
  try {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (keyword) params.set("keyword", keyword);
    if (scopeType) params.set("scope_type", scopeType);
    if (scopeID) params.set("scope_id", scopeID);
    const topics = await apiFetch(`/forum/topics?${params.toString()}`, { method: "GET" });
    const topicList = topics.list || [];
    const scopeHint = scopeType ? `Current scope: ${forumScopeLabel(scopeType, scopeID || null)}` : "General discussion across contests, problems, and training topics.";

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">Forum</h1>
          <p class="view-subtitle">${escapeHTML(scopeHint)}</p>
        </div>
      </div>
      ${renderForumSummaryCards(topicList, scopeType)}
      <section class="detail-grid" style="margin-top:18px; align-items:start;">
        <article class="detail-card forum-panel-card">
          <div class="view-header compact">
            <div>
              <h3>Browse Topics</h3>
              <p class="view-subtitle">Filter discussion by scope or keyword.</p>
            </div>
            <div class="pill-row forum-scope-tabs">
              <a class="ghost-button ${!scopeType ? "active-filter" : ""}" href="#/forum">All</a>
              <a class="ghost-button ${scopeType === "general" ? "active-filter" : ""}" href="#/forum?scope_type=general">General</a>
              <a class="ghost-button ${scopeType === "problem" ? "active-filter" : ""}" href="#/forum?scope_type=problem${scopeID ? `&scope_id=${scopeID}` : ""}">Problem</a>
              <a class="ghost-button ${scopeType === "contest" ? "active-filter" : ""}" href="#/forum?scope_type=contest${scopeID ? `&scope_id=${scopeID}` : ""}">Contest</a>
            </div>
          </div>
          <form id="forum-filter-form" class="forum-filter-form">
            <label class="field-label">Keyword</label>
            <input class="text-input" name="keyword" value="${escapeHTML(keyword)}" placeholder="Search title or content" />
            <label class="field-label">Scope Type</label>
            <select class="select-input" name="scope_type">
              <option value="" ${!scopeType ? "selected" : ""}>all</option>
              <option value="general" ${scopeType === "general" ? "selected" : ""}>general</option>
              <option value="problem" ${scopeType === "problem" ? "selected" : ""}>problem</option>
              <option value="contest" ${scopeType === "contest" ? "selected" : ""}>contest</option>
            </select>
            <label class="field-label">Scope ID</label>
            <input class="text-input" name="scope_id" value="${escapeHTML(scopeID)}" placeholder="optional" />
            <div class="forum-filter-actions">
              <button class="primary-button" type="submit">Apply</button>
              <a class="ghost-button" href="#/forum">Reset</a>
            </div>
          </form>
        </article>
        <aside class="detail-card forum-panel-card">
          <div class="view-header compact">
            <div>
              <h3>New Topic</h3>
              <p class="view-subtitle">Start a thread for problem ideas, contest clarifications, or general training discussion.</p>
            </div>
          </div>
          ${state.user ? `
            <form id="forum-create-form">
              <label class="field-label">Title</label>
              <input class="text-input" name="title" required />
              <label class="field-label">Scope Type</label>
              <select class="select-input" name="scope_type">
                <option value="general" ${scopeType === "general" || !scopeType ? "selected" : ""}>general</option>
                <option value="problem" ${scopeType === "problem" ? "selected" : ""}>problem</option>
                <option value="contest" ${scopeType === "contest" ? "selected" : ""}>contest</option>
              </select>
              <label class="field-label">Scope ID</label>
              <input class="text-input" name="scope_id" value="${escapeHTML(scopeID)}" placeholder="required for problem/contest" />
              <label class="field-label">Content</label>
              <textarea class="text-area" name="content" rows="10" required></textarea>
              <div class="view-subtitle" style="margin-top:8px;">Markdown is supported. Use fenced code blocks for snippets and keep clarifications concise.</div>
              <div style="margin-top:14px;"><button class="primary-button" type="submit">Post Topic</button></div>
            </form>
          ` : `<p class="view-subtitle">Login to create new topics and participate in discussion.</p>`}
        </aside>
      </section>
      <section style="margin-top:18px;">${renderForumTopicCards(topicList)}</section>
    `;

    document.getElementById("forum-filter-form")?.addEventListener("submit", (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const next = new URLSearchParams();
      const nextKeyword = (form.get("keyword") || "").toString().trim();
      const nextScopeType = (form.get("scope_type") || "").toString().trim();
      const nextScopeID = (form.get("scope_id") || "").toString().trim();
      if (nextKeyword) next.set("keyword", nextKeyword);
      if (nextScopeType) next.set("scope_type", nextScopeType);
      if (nextScopeID) next.set("scope_id", nextScopeID);
      next.set("page", "1");
      next.set("page_size", String(pageSize));
      location.hash = `#/forum?${next.toString()}`;
    });

    document.getElementById("forum-create-form")?.addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const createScopeType = (form.get("scope_type") || "general").toString();
      const rawScopeID = (form.get("scope_id") || "").toString().trim();
      try {
        const topic = await apiFetch("/forum/topics", {
          method: "POST",
          body: JSON.stringify({
            title: form.get("title"),
            content: form.get("content"),
            scope_type: createScopeType,
            scope_id: createScopeType === "general" || !rawScopeID ? null : Number(rawScopeID),
          }),
        });
        setFlash(`Topic #${topic.id} created`, false);
        location.hash = `#/forum/topics/${topic.id}`;
      } catch (err) {
        setFlash(err.message, true);
      }
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load forum failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

async function renderForumTopicDetail(id) {
  app.innerHTML = `<div class="detail-card"><p>Loading topic...</p></div>`;
  try {
    const detail = await apiFetch(`/forum/topics/${id}`, { method: "GET" });
    const canManageTopic = canManageForumContent(detail.author_id);
    const canModerate = canModerateForum();

    app.innerHTML = `
      <div class="view-header">
        <div>
          <h1 class="view-title">${escapeHTML(detail.title)}</h1>
          <p class="view-subtitle">${escapeHTML(detail.author_name)} · ${escapeHTML(detail.created_at)} · ${escapeHTML(forumScopeLabel(detail.scope_type, detail.scope_id))}</p>
        </div>
        <div style="display:flex; gap:10px; flex-wrap:wrap;">
          <a class="ghost-button" href="#/forum?scope_type=${encodeURIComponent(detail.scope_type)}${detail.scope_id ? `&scope_id=${detail.scope_id}` : ""}">Back to Topics</a>
          <a class="ghost-button" href="${forumScopeLink(detail.scope_type, detail.scope_id)}">Open Scope</a>
          ${canManageTopic ? `<button class="ghost-button" type="button" id="forum-topic-edit-btn">Edit Topic</button><button class="ghost-button" type="button" id="forum-topic-delete-btn">Delete Topic</button>` : ""}
        </div>
      </div>
      <section class="detail-grid" style="align-items:start;">
        <article class="detail-card forum-topic-detail-main">
          <div class="forum-topic-state-bar">
            ${forumStatusPills(detail)}
          </div>
          ${renderMarkdownBlock(detail.content || "")}
        </article>
        <aside class="detail-card forum-topic-sidebar">
          <h3>Topic Meta</h3>
          <div class="teaching-meta-grid">
            <div class="teaching-meta-card"><span class="teaching-meta-label">Scope</span><span class="teaching-meta-value">${escapeHTML(forumScopeLabel(detail.scope_type, detail.scope_id))}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Replies</span><span class="teaching-meta-value">${detail.reply_count}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Created</span><span class="teaching-meta-value mono">${escapeHTML(detail.created_at)}</span></div>
            <div class="teaching-meta-card"><span class="teaching-meta-label">Updated</span><span class="teaching-meta-value mono">${escapeHTML(detail.updated_at)}</span></div>
          </div>
          ${canModerate ? `
            <div class="forum-moderation-box">
              <h4 style="margin:0 0 10px;">Moderation</h4>
              <div class="forum-topic-footer">
                <button class="ghost-button" type="button" id="forum-toggle-pin-btn">${detail.is_pinned ? "Unpin" : "Pin"}</button>
                <button class="ghost-button" type="button" id="forum-toggle-lock-btn">${detail.is_locked ? "Unlock" : "Lock"}</button>
              </div>
            </div>
          ` : ""}
        </aside>
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <div class="view-header compact">
          <div>
            <h3>Replies</h3>
            <p class="view-subtitle">${detail.replies?.length || 0} reply(s)</p>
          </div>
          ${detail.is_locked ? `<span class="status-pill status-pending">Locked</span>` : ""}
        </div>
        ${(detail.replies || []).length ? `
          <div class="forum-reply-list">
            ${(detail.replies || []).map((item) => `
              <article class="forum-reply-card">
                <div class="view-header compact">
                  <div>
                    <strong>${escapeHTML(item.author_name)}</strong>
                    <p class="view-subtitle">${escapeHTML(item.created_at)}</p>
                  </div>
                  ${canManageForumContent(item.author_id) ? `<div style="display:flex; gap:8px; flex-wrap:wrap;"><button class="ghost-button forum-reply-edit" type="button" data-reply-id="${item.id}" data-reply-content="${encodeURIComponent(item.content || "")}">Edit</button><button class="ghost-button forum-reply-delete" type="button" data-reply-id="${item.id}">Delete</button></div>` : ""}
                </div>
                ${renderMarkdownBlock(item.content || "")}
              </article>
            `).join("")}
          </div>
        ` : `<p class="view-subtitle">No replies yet.</p>`}
      </section>
      <section class="detail-card" style="margin-top:18px;">
        <h3>Reply</h3>
        ${state.user ? `
          ${detail.is_locked && !canModerate ? `<p class="view-subtitle">This topic is locked. Only moderators can reply.</p>` : `
            <form id="forum-reply-form">
              <label class="field-label">Content</label>
              <textarea class="text-area" name="content" rows="8" required></textarea>
              <div class="view-subtitle" style="margin-top:8px;">Markdown is supported in replies.</div>
              <div style="margin-top:14px;"><button class="primary-button" type="submit">Post Reply</button></div>
            </form>
          `}
        ` : `<p class="view-subtitle">Login to reply.</p>`}
      </section>
    `;

    document.getElementById("forum-reply-form")?.addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await apiFetch(`/forum/topics/${id}/replies`, {
          method: "POST",
          body: JSON.stringify({ content: form.get("content") }),
        });
        setFlash("Reply posted", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-topic-edit-btn")?.addEventListener("click", async () => {
      const nextTitle = window.prompt("Edit topic title", detail.title);
      if (nextTitle === null) return;
      const nextContent = window.prompt("Edit topic content", detail.content || "");
      if (nextContent === null) return;
      try {
        await apiFetch(`/forum/topics/${id}`, {
          method: "PUT",
          body: JSON.stringify({ title: nextTitle, content: nextContent }),
        });
        setFlash("Topic updated", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-topic-delete-btn")?.addEventListener("click", async () => {
      if (!window.confirm(`Delete topic #${id}?`)) return;
      try {
        await apiFetch(`/forum/topics/${id}`, { method: "DELETE" });
        setFlash("Topic deleted", false);
        location.hash = "#/forum";
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-toggle-pin-btn")?.addEventListener("click", async () => {
      try {
        await apiFetch(`/forum/topics/${id}`, {
          method: "PUT",
          body: JSON.stringify({ title: detail.title, content: detail.content, is_pinned: !detail.is_pinned, is_locked: detail.is_locked }),
        });
        setFlash(detail.is_pinned ? "Topic unpinned" : "Topic pinned", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.getElementById("forum-toggle-lock-btn")?.addEventListener("click", async () => {
      try {
        await apiFetch(`/forum/topics/${id}`, {
          method: "PUT",
          body: JSON.stringify({ title: detail.title, content: detail.content, is_pinned: detail.is_pinned, is_locked: !detail.is_locked }),
        });
        setFlash(detail.is_locked ? "Topic unlocked" : "Topic locked", false);
        return renderForumTopicDetail(id);
      } catch (err) {
        setFlash(err.message, true);
      }
    });

    document.querySelectorAll(".forum-reply-edit").forEach((button) => {
      button.addEventListener("click", async () => {
        const replyID = button.dataset.replyId;
        const current = decodeURIComponent(button.dataset.replyContent || "");
        const nextContent = window.prompt("Edit reply", current);
        if (nextContent === null) return;
        try {
          await apiFetch(`/forum/replies/${replyID}`, {
            method: "PUT",
            body: JSON.stringify({ content: nextContent }),
          });
          setFlash(`Reply #${replyID} updated`, false);
          return renderForumTopicDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });

    document.querySelectorAll(".forum-reply-delete").forEach((button) => {
      button.addEventListener("click", async () => {
        const replyID = button.dataset.replyId;
        if (!window.confirm(`Delete reply #${replyID}?`)) return;
        try {
          await apiFetch(`/forum/replies/${replyID}`, { method: "DELETE" });
          setFlash(`Reply #${replyID} deleted`, false);
          return renderForumTopicDetail(id);
        } catch (err) {
          setFlash(err.message, true);
        }
      });
    });
  } catch (err) {
    app.innerHTML = `<div class="detail-card"><p>Load forum topic failed: ${escapeHTML(err.message)}</p></div>`;
  }
}

