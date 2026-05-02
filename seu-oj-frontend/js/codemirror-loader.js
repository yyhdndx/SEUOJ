import { basicSetup } from "codemirror";
import { autocompletion, completeFromList } from "@codemirror/autocomplete";
import { indentWithTab } from "@codemirror/commands";
import { Compartment, EditorState } from "@codemirror/state";
import { HighlightStyle, indentUnit, syntaxHighlighting } from "@codemirror/language";
import { EditorView, keymap } from "@codemirror/view";
import { cpp } from "@codemirror/lang-cpp";
import { go } from "@codemirror/lang-go";
import { java } from "@codemirror/lang-java";
import { markdown } from "@codemirror/lang-markdown";
import { python } from "@codemirror/lang-python";
import { rust } from "@codemirror/lang-rust";
import { tags } from "@lezer/highlight";

const completionItems = {
  cpp: [
    "auto", "bool", "break", "case", "catch", "cin", "class", "const", "continue", "cout",
    "else", "for", "if", "int", "long", "namespace", "return", "std", "struct", "switch",
    "template", "try", "using", "vector", "void", "while",
  ],
  c: [
    "break", "case", "char", "const", "continue", "double", "else", "enum", "float", "for",
    "if", "int", "long", "printf", "return", "scanf", "sizeof", "static", "struct", "switch",
    "typedef", "void", "while",
  ],
  python3: [
    "and", "as", "break", "class", "continue", "def", "elif", "else", "False", "for", "from",
    "if", "import", "in", "lambda", "len", "None", "not", "or", "print", "range", "return",
    "True", "while",
  ],
  java: [
    "ArrayList", "class", "else", "extends", "for", "HashMap", "if", "import", "int", "long",
    "new", "private", "public", "return", "Scanner", "static", "String", "System", "void",
    "while",
  ],
  go: [
    "append", "break", "case", "const", "continue", "defer", "else", "error", "fallthrough",
    "fmt", "for", "func", "go", "if", "import", "interface", "make", "map", "package", "range",
    "return", "select", "struct", "switch", "type", "var",
  ],
  rust: [
    "break", "const", "else", "enum", "fn", "for", "if", "impl", "let", "loop", "match", "mod",
    "mut", "Option", "pub", "Result", "return", "Self", "Some", "String", "struct", "trait",
    "use", "vec!", "while",
  ],
};

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
    lineHeight: "1.65",
  },
  ".cm-content": {
    minHeight: "100%",
    padding: "14px 0",
    caretColor: "#8c4b12",
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
  ".cm-activeLine": {
    backgroundColor: "rgba(196, 167, 104, 0.12)",
  },
  ".cm-activeLineGutter": {
    backgroundColor: "rgba(196, 167, 104, 0.18)",
  },
  ".cm-selectionBackground, ::selection": {
    backgroundColor: "rgba(176, 139, 71, 0.24)",
  },
  ".cm-tooltip": {
    border: "1px solid #d7c49c",
    backgroundColor: "#fffaf0",
    boxShadow: "0 12px 26px rgba(47, 36, 12, 0.16)",
  },
  ".cm-tooltip-autocomplete ul li[aria-selected]": {
    backgroundColor: "rgba(176, 139, 71, 0.16)",
    color: "#3b2a13",
  },
  ".cm-focused": {
    outline: "none",
  },
  "&.cm-focused": {
    borderColor: "#b08b47",
    boxShadow: "0 0 0 2px rgba(176, 139, 71, 0.16)",
  },
  ".cm-matchingBracket": {
    backgroundColor: "rgba(45, 91, 159, 0.18)",
    color: "#204d83",
  },
});

const problemEditorHighlightStyle = HighlightStyle.define([
  { tag: [tags.keyword, tags.modifier], color: "#9d3f18", fontWeight: "600" },
  { tag: [tags.string, tags.special(tags.string)], color: "#2d7d46" },
  { tag: [tags.number, tags.integer, tags.float, tags.bool], color: "#155d98" },
  { tag: [tags.comment, tags.lineComment, tags.blockComment], color: "#8a7d68", fontStyle: "italic" },
  { tag: [tags.function(tags.variableName), tags.labelName], color: "#8c4b12" },
  { tag: [tags.className, tags.typeName], color: "#6b3fa0" },
  { tag: [tags.operator, tags.punctuation, tags.separator], color: "#5a4f3c" },
  { tag: [tags.variableName, tags.propertyName], color: "#1f252f" },
]);

const solutionMarkdownEditorTheme = EditorView.theme({
  "&": {
    height: "100%",
    minHeight: "460px",
    fontSize: "14px",
    color: "#1f252f",
    backgroundColor: "#fbfaf6",
    border: "1px solid #c7b894",
    borderRadius: "8px",
  },
  ".cm-scroller": {
    overflow: "auto",
    fontFamily: '"IBM Plex Mono", "Consolas", monospace',
    lineHeight: "1.7",
  },
  ".cm-content": {
    minHeight: "100%",
    padding: "14px 0",
    caretColor: "#8c4b12",
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
  ".cm-activeLine": {
    backgroundColor: "rgba(196, 167, 104, 0.12)",
  },
  ".cm-activeLineGutter": {
    backgroundColor: "rgba(196, 167, 104, 0.18)",
  },
  ".cm-focused": {
    outline: "none",
  },
  "&.cm-focused": {
    borderColor: "#b08b47",
    boxShadow: "0 0 0 2px rgba(176, 139, 71, 0.16)",
  },
});

function normalizeEditorLanguage(language) {
  if (language === "c") {
    return "c";
  }
  if (completionItems[language]) {
    return language;
  }
  return "cpp";
}

function getLanguageSupport(language) {
  const normalized = normalizeEditorLanguage(language);
  switch (normalized) {
    case "python3":
      return python();
    case "java":
      return java();
    case "go":
      return go();
    case "rust":
      return rust();
    case "c":
    case "cpp":
    default:
      return cpp();
  }
}

function getCompletionExtension(language) {
  const normalized = normalizeEditorLanguage(language);
  const items = (completionItems[normalized] || []).map((label) => ({ label, type: "keyword" }));
  if (!items.length) {
    return [];
  }
  const support = getLanguageSupport(language);
  return support.language.data.of({
    autocomplete: completeFromList(items),
  });
}

function getIndentExtension(size) {
  const indentSize = size === 2 ? 2 : 4;
  return [
    EditorState.tabSize.of(indentSize),
    indentUnit.of(" ".repeat(indentSize)),
  ];
}

function createProblemCodeEditor(textarea, options = {}) {
  if (!textarea) {
    return null;
  }

  const initialLanguage = options.language || textarea.dataset.language || "cpp";
  const initialIndentSize = Number(options.indentSize || textarea.dataset.indentSize || 4) || 4;
  const languageCompartment = new Compartment();
  const indentCompartment = new Compartment();
  const completionCompartment = new Compartment();

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
        keymap.of([indentWithTab]),
        autocompletion({ activateOnTyping: true }),
        syntaxHighlighting(problemEditorHighlightStyle),
        problemEditorTheme,
        languageCompartment.of(getLanguageSupport(initialLanguage)),
        completionCompartment.of(getCompletionExtension(initialLanguage)),
        indentCompartment.of(getIndentExtension(initialIndentSize)),
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
    setLanguage(language) {
      textarea.dataset.language = language;
      view.dispatch({
        effects: [
          languageCompartment.reconfigure(getLanguageSupport(language)),
          completionCompartment.reconfigure(getCompletionExtension(language)),
        ],
      });
    },
    setIndentSize(size) {
      const indentSize = size === 2 ? 2 : 4;
      textarea.dataset.indentSize = String(indentSize);
      view.dispatch({
        effects: indentCompartment.reconfigure(getIndentExtension(indentSize)),
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

function createSolutionMarkdownEditor(textarea, options = {}) {
  if (!textarea) {
    return null;
  }

  const host = document.createElement("div");
  host.className = "solution-markdown-editor-host";
  textarea.insertAdjacentElement("beforebegin", host);
  textarea.classList.add("is-codemirror-hidden");

  const syncTextarea = EditorView.updateListener.of((update) => {
    if (!update.docChanged) return;
    const value = update.state.doc.toString();
    textarea.value = value;
    if (typeof options.onChange === "function") {
      options.onChange(value);
    }
  });

  const view = new EditorView({
    state: EditorState.create({
      doc: textarea.value,
      extensions: [
        basicSetup,
        keymap.of([indentWithTab]),
        markdown(),
        solutionMarkdownEditorTheme,
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
      if (typeof options.onChange === "function") {
        options.onChange(next);
      }
    },
    destroy() {
      view.destroy();
      host.remove();
      textarea.classList.remove("is-codemirror-hidden");
    },
  };
}

window.createSolutionMarkdownEditor = createSolutionMarkdownEditor;
