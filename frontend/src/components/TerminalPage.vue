<template>
    <div class="terminal-page">
      <div ref="terminalContainer" class="terminal-container"></div>
    </div>
  </template>
  
  <script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import '@xterm/xterm/css/xterm.css';
  
  const terminalContainer = ref<HTMLElement | null>(null);
  let socket: WebSocket;
  
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
  
      // 连接 WebSocket
      socket = new WebSocket('ws://localhost:8080/ws');
  
      socket.onopen = () => {
        terminal.writeln('Connected to remote server.');
      };
  
      socket.onmessage = (event) => {
        terminal.write(event.data);
      };
  
      socket.onclose = () => {
        terminal.writeln('Disconnected from remote server.');
      };
  
      // 处理用户输入
      terminal.onData((data) => {
        console.log('User input:', data);
        socket.send(data);
      });
  
      // 处理窗口大小变化
      window.addEventListener('resize', () => {
        fitAddon.fit();
      });
    }
  });
  </script>
  
  <style scoped>
  .terminal-page {
    width: 100%;
    height: 100vh;
    padding: 10px;
    box-sizing: border-box;
  }
  .terminal-container {
    width: 100%;
    height: 100%;
  }
  </style>