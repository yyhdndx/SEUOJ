import { basicSetup, EditorView } from "codemirror";
import { EditorState } from "@codemirror/state";

const problemEditorTheme = EditorView.theme({
  "&": {
    height: "100%",
    fontSize: "14px",
    color: "#1f252f",
    backgroundColor: "#fbfaf6",
    border: "1px solid #c7b894",
    borderRadius: "8px",
  },
  ".cm-scroller": {
    overflow: "auto",
    fontFamily: '"IBM Plex Mono", "Consolas", monospace',
    lineHeight: "1.6",
  },
  ".cm-content": {
    minHeight: "100%",
    padding: "14px 0",
  },
  ".cm-line": {
    padding: "0 16px",
  },
  ".cm-gutters": {
    backgroundColor: "#f4efe4",
    color: "#7a705a",
    borderRight: "1px solid #dacfb6",
    borderTopLeftRadius: "8px",
    borderBottomLeftRadius: "8px",
  },
  ".cm-activeLine, .cm-activeLineGutter": {
    backgroundColor: "rgba(196, 167, 104, 0.12)",
  },
  ".cm-focused": {
    outline: "none",
  },
  "&.cm-focused": {
    borderColor: "#b08b47",
    boxShadow: "0 0 0 2px rgba(176, 139, 71, 0.16)",
  },
});

function createProblemCodeEditor(textarea) {
  if (!textarea) {
    return null;
  }

  const host = document.createElement("div");
  host.className = "problem-code-editor-host";
  textarea.insertAdjacentElement("beforebegin", host);
  textarea.classList.add("is-codemirror-hidden");

  const syncTextarea = EditorView.updateListener.of((update) => {
    if (update.docChanged) {
      textarea.value = update.state.doc.toString();
    }
  });

  const view = new EditorView({
    state: EditorState.create({
      doc: textarea.value,
      extensions: [
        basicSetup,
        EditorState.tabSize.of(4),
        problemEditorTheme,
        syncTextarea,
      ],
    }),
    parent: host,
  });

  return {
    focus() {
      view.focus();
    },
    getValue() {
      return view.state.doc.toString();
    },
    setValue(value) {
      const next = value ?? "";
      textarea.value = next;
      view.dispatch({
        changes: {
          from: 0,
          to: view.state.doc.length,
          insert: next,
        },
      });
    },
    destroy() {
      view.destroy();
      host.remove();
      textarea.classList.remove("is-codemirror-hidden");
    },
  };
}

window.createProblemCodeEditor = createProblemCodeEditor;
