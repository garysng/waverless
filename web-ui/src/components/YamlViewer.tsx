import { useRef } from 'react';
import Editor from '@monaco-editor/react';
import type { editor } from 'monaco-editor';

interface YamlViewerProps {
  yaml: string;
  height?: string;
}

const YamlViewer = ({ yaml, height = 'calc(100vh - 250px)' }: YamlViewerProps) => {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);

  const handleEditorDidMount = (editor: editor.IStandaloneCodeEditor) => {
    editorRef.current = editor;
    // Content is expanded by default, no need to fold
  };

  return (
    <div style={{ height, width: '100%' }}>
      <Editor
        height="100%"
        language="yaml"
        value={yaml}
        theme="vs-light"
        options={{
          readOnly: true,
          minimap: { enabled: true },
          scrollBeyondLastLine: false,
          fontSize: 13,
          lineNumbers: 'on',
          folding: true,
          foldingStrategy: 'indentation',
          showFoldingControls: 'always',
          automaticLayout: true,
          wordWrap: 'off',
          renderWhitespace: 'selection',
          scrollbar: {
            vertical: 'visible',
            horizontal: 'visible',
            useShadows: false,
            verticalScrollbarSize: 10,
            horizontalScrollbarSize: 10,
          },
        }}
        onMount={handleEditorDidMount}
      />
    </div>
  );
};

export default YamlViewer;
