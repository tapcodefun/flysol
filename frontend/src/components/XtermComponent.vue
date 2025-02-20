<template>
    <div ref="terminalContainer" class="terminal-container"></div>
  </template>
  
  <script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import '@xterm/xterm/css/xterm.css';
  
  const terminalContainer = ref<HTMLElement | null>(null);
  
  onMounted(() => {
    if (terminalContainer.value) {
      const terminal = new Terminal({
        cursorBlink: true,
        theme: {
          background: '#1e1e1e',
          foreground: '#ffffff',
        },
      });
  
      const fitAddon = new FitAddon();
      terminal.loadAddon(fitAddon);
  
      terminal.open(terminalContainer.value);
      fitAddon.fit();
  
      terminal.writeln('Welcome to Xterm.js in Vue 3 + TypeScript!');
  
      // 处理窗口大小变化
      window.addEventListener('resize', () => {
        fitAddon.fit();
      });
  
      // 处理用户输入
      terminal.onData((data) => {
        terminal.write(data);
      });
    }
  });
  </script>
  
  <style scoped>
  .terminal-container {
    width: 100%;
    height: 100%;
    padding: 10px;
    box-sizing: border-box;
  }
  </style>