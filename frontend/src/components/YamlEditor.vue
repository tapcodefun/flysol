<template>
    <div ref="editorContainer" class="yaml-editor-container"></div>
  </template>
  
  <script lang="ts" setup>
  import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
  import { EditorState } from '@codemirror/state';
  import { EditorView, keymap } from '@codemirror/view';
  import { yaml } from '@codemirror/lang-yaml';
  import { basicSetup } from 'codemirror';
  import { oneDark } from '@codemirror/theme-one-dark';
  
  const props = defineProps({
    modelValue: String,
  });
  
  const emit = defineEmits(['update:modelValue']);
  
  const editorContainer = ref<HTMLElement | null>(null);
  let editorView: EditorView | null = null;
  
  // 初始化编辑器
  const initEditor = () => {
    if (!editorContainer.value) return;
  
    const startState = EditorState.create({
      doc: props.modelValue,
      extensions: [
        basicSetup,
        yaml(),
        oneDark, // 使用暗色主题
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            const content = update.state.doc.toString();
            emit('update:modelValue', content);
          }
        }),
        keymap.of([
          {
            key: 'Mod-s', // 支持 Ctrl+S / Cmd+S 保存
            preventDefault: true,
            run: () => {
              console.log('Save triggered');
              return true;
            },
          },
        ]),
      ],
    });
  
    editorView = new EditorView({
      state: startState,
      parent: editorContainer.value,
    });
  };
  
  // 监听外部 modelValue 的变化
  watch(
    () => props.modelValue,
    (newValue) => {
      if (editorView && newValue !== editorView.state.doc.toString()) {
        editorView.dispatch({
          changes: {
            from: 0,
            to: editorView.state.doc.length,
            insert: newValue,
          },
        });
      }
    }
  );
  
  // 组件挂载时初始化编辑器
  onMounted(() => {
    initEditor();
  });
  
  // 组件销毁时销毁编辑器
  onBeforeUnmount(() => {
    if (editorView) {
      editorView.destroy();
    }
  });
  </script>
  
  <style scoped>
  .yaml-editor-container {
    height: 100%;
    overflow: hidden;
    background: #1e1e1e;
  }
  
  .yaml-editor-container :deep(.cm-editor) {
    height: 100%;
  }
  
  .yaml-editor-container :deep(.cm-scroller) {
    overflow: auto;
  }
  
  .yaml-editor-container :deep(.cm-gutters) {
    background: #1e1e1e;
    border-right: 1px solid #333;
  }
  
  .yaml-editor-container :deep(.cm-content) {
    font-family: 'Consolas', monospace;
    font-size: 14px;
    line-height: 1.5;
    color: #ffffff;
  }
  
  .yaml-editor-container :deep(.cm-activeLine) {
    background: #2a2a2a;
  }
  
  .yaml-editor-container :deep(.cm-activeLineGutter) {
    background: #2a2a2a;
  }
  
  /* 自定义滚动条样式 */
  .yaml-editor-container :deep(.cm-scroller)::-webkit-scrollbar {
    width: 10px;
    height: 10px;
  }
  
  .yaml-editor-container :deep(.cm-scroller)::-webkit-scrollbar-track {
    background: #1e1e1e;
  }
  
  .yaml-editor-container :deep(.cm-scroller)::-webkit-scrollbar-thumb {
    background: #666;
    border-radius: 5px;
  }
  
  .yaml-editor-container :deep(.cm-scroller)::-webkit-scrollbar-thumb:hover {
    background: #888;
  }
  
  .yaml-editor-container :deep(.cm-scroller)::-webkit-scrollbar-corner {
    background: #1e1e1e;
  }
  </style>