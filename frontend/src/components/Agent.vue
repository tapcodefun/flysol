<template>
    <div class="terminal-page">
      <div class="button-container">
        <button @click="sendCommand('cd /home && chmod +x agent.sh && ./agent.sh')" class="cmd-button">安装与升级</button>
        <button v-if="screens.length==0" @click="sendCommand('cd /home/bot && screen -S bot')" class="cmd-button">创建屏幕</button>
        <div v-for="scr in screens" :key="scr" class="screen-container">
          <button @click="sendCommand('cd /home/bot && screen -r ' + scr)" class="cmd-button">
            屏幕{{ scr }}
          </button>
          <span @click="closeScreen(scr)" class="close-button">×</span>
        </div>
        <button @click="sendCommand('cd /home/bot && ./run.sh')" class="cmd-button">运行程序</button>
        <button @click="sendCommand('Ctrl+C')" class="cmd-button">终止运行</button>
        <button @click="sendCommand('cd /home/bot && ./kill-process.sh')" class="cmd-button">强制停止</button>
        <button @click="sendCommand('Ctrl+A+D')" class="cmd-button">退出屏幕</button>
        <button @click="saveConfig()" class="cmd-button">保存参数</button>
        <button @click="toggleView" class="cmd-button">
          {{ showEditor ? '显示命令行' : '显示文件编辑' }}
        </button>
        <button class="cmd-button" @click="$emit('close')">关闭</button>
        <div class="host-display">当前主机: {{ hostname }}</div>
      </div>
      <div class="content-container">
        <div :class="['editor-container', { 'hidden': !showEditor }]">
          <YamlEditor v-model="yamlContent" />
        </div>
        <div ref="terminalContainer" :class="['terminal-container', { 'hidden': showEditor }]"></div>
      </div>
    </div>
  </template>
  
  <script lang="ts" setup>
  import YamlEditor from './YamlEditor.vue';
  import { onMounted, ref, computed } from 'vue';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import '@xterm/xterm/css/xterm.css';
  import { defineProps } from 'vue';
  
  const terminalContainer = ref<HTMLElement | null>(null);
  const socket = ref<WebSocket | null>(null);
  const terminal = ref<Terminal | null>(null);
  const yamlContent = ref('');
  const screens = ref([]);
  const showEditor = ref(false);
  const props = defineProps({
    host: {
      type: String,
      required: true
    },
    token: {
      required: true
    },
    hostname: {
      type: String,
      default: ''
    }
  });
  const host = computed(() => props.host);
  const token = computed(() => props.token);
  const hostname = computed(() => props.hostname);
  const hosturl = `http://${host.value}:5189?model=run&code=${token.value}`
  // const props = defineProps({
  //   hosturl:{
  //     type:String,
  //     required:true
  //   }
  // });
  // const hosturl = computed(() => props.hosturl);
  const saveConfig = async () => {
    try {
      const response = await fetch(hosturl+'/config', {
        method: 'POST', // 指定请求方法为 POST
        headers: {
          'Content-Type': 'application/json', // 设置请求头，指定发送的数据格式为 JSON
        },
        body: JSON.stringify({file:'',content:yamlContent.value}), // 将配置数据转换为 JSON 字符串并放入请求体
      });
  
      if (response.ok) {
        toggleView()
        terminal.value?.writeln('Config saved successfully');
      } else {
        console.error('Failed to save config:', response.statusText);
      }
    } catch (error) {
      console.error('Error saving config:', error);
    }
  };
  
  const closeScreen = async (scr:string) => {
    showEditor.value = false;
    if (socket.value?.readyState === WebSocket.OPEN && terminal.value) {
      var command = 'cd /home/bot && screen -X -S '+scr+' quit'
      terminal.value.writeln('\n' + command);
      socket.value.send(command + "\n");
      setTimeout(() => {
        fetchScreen();
      }, 5000); // 5000 毫秒 = 5 秒
    } else {
      console.error('WebSocket not connected');
      terminal.value?.writeln('\x1b[31mError: Not connected to server\x1b[0m');
    }
  }
  const sendCommand = (command: string) => {
    showEditor.value = false;
    if (socket.value?.readyState === WebSocket.OPEN && terminal.value) {
      terminal.value.writeln('\n' + command);
      socket.value.send(command + "\n");
      if(command=='cd /home/bot && screen -S bot'){
        setTimeout(() => {
          fetchScreen();
        }, 5000); // 5000 毫秒 = 5 秒
      }
    } else {
      console.error('WebSocket2 not connected');
      terminal.value?.writeln('\x1b[31mError: Not connected to server\x1b[0m');
    }
  };
  
  const toggleView = () => {
    showEditor.value = !showEditor.value;
  };
  
  const fetchConfig = async () => {
    try {
      const response = await fetch(`http://${host.value}:5189/config?code=${token.value}`, {
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const content = await response.json();
        yamlContent.value = content.message;
      } else {
        console.error('Failed to fetch config:', response.statusText);
      }
    } catch (error) {
      console.error('Error fetching config:', error);
    }
  };
  
  const fetchScreen = async () => {
    try {
      const response = await fetch(`http://${host.value}:5189/screen?code=${token.value}`, {
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const content = await response.json();
        if(content.error){
          console.log(content.error);
        }else{
          screens.value = content.screens;
        }
      } else {
        screens.value = [];
        console.error('Failed to fetch config:', response.statusText);
      }
    } catch (error) {
      console.error('Error fetching config:', error);
    }
  };
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5; // 最大重连次数
  const reconnectDelay = 3000; // 重连延迟时间（毫秒）
  
  function connectWebSocket() {
    const url = new URL(window.location.href);
    const urlmodel = url.searchParams.get('model') || 'run';
    // 从hosturl创建URL对象以确保正确的URL格式
    const baseUrl = new URL(hosturl);
    // 使用URL对象的protocol来确定WebSocket协议
    const wsProtocol = baseUrl.protocol === 'https:' ? 'wss:' : 'ws:';
    // 构建WebSocket URL
    const socketUrl = `${wsProtocol}//${baseUrl.host}/ws${urlmodel}`;
    // const socketUrl = `ws://${host}:5189/wsrun`
    // const socketUrl = `ws://${host.value}:5189/ws${urlmodel}?code=${token.value}`;
    console.log(socketUrl);
  
    socket.value = new WebSocket(socketUrl);
  
    socket.value.onopen = () => {
      console.log('WebSocket connected');
      terminal.value?.writeln('Connected to remote server.');
      reconnectAttempts = 0; // 重置重连次数
      setTimeout(() => {
        fetchScreen();
      }, 5000); // 5000 毫秒 = 5 秒
    };
  
    socket.value.onmessage = (event) => {
      let data = event.data.replace(/[\b\r]/g, '');
      const lines = data.split('\n');
      if (lines.length == 1) {
        terminal.value?.writeln(lines[0]);
      } else {
        lines.forEach((line: string, index: number) => {
          if (line) {
            index === lines.length - 1
              ? terminal.value?.write(line)
              : terminal.value?.writeln(line);
          }
        });
      }
    };
  
    socket.value.onclose = () => {
      terminal.value?.writeln('Disconnected from remote server.');
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        terminal.value?.writeln(`Attempting to reconnect (${reconnectAttempts}/${maxReconnectAttempts})...`);
        setTimeout(connectWebSocket, reconnectDelay);
      } else {
        terminal.value?.writeln('Max reconnection attempts reached. Please check your connection.');
      }
    };
  
    socket.value.onerror = () => {
      terminal.value?.writeln('\x1b[31mWebSocket connection error\x1b[0m');
    };
  }
  
  onMounted(() => {
    if (terminalContainer.value) {
      terminal.value = new Terminal({
        cursorBlink: true,
        theme: {
          background: '#1e1e1e',
          foreground: '#ffffff',
        },
        fontSize: 14,
        fontFamily: 'Consolas, monospace',
        lineHeight: 1.2,
        allowProposedApi: true,
        convertEol: true,
      });
  
      const fitAddon = new FitAddon();
      terminal.value.loadAddon(fitAddon);
      terminal.value.open(terminalContainer.value);
      fitAddon.fit();
  
      connectWebSocket(); // 初始连接
    }
    fetchConfig();
    fetchScreen();
  });
  </script>
  
  <style scoped>
  .terminal-page {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: #1e1e1e;
  }
  
  .button-container {
    padding: 12px;
    background: #252526;
    border-bottom: 1px solid #333;
    display: flex;
    gap: 10px;
    flex-shrink: 0;
  }
  
  .cmd-button {
    background: #3c3c3c;
    border: 1px solid #4d4d4d;
    color: #d4d4d4;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    transition: all 0.2s ease;
  }
  
  .screen-container {
    position: relative;
    display: inline-block;
    margin: 0px;
  }
  
  .close-button {
    position: absolute;
    top: -10px;
    right: -10px;
    background-color: red;
    color: white;
    border-radius: 50%;
    width: 20px;
    height: 20px;
    text-align: center;
    line-height: 20px;
    cursor: pointer;
    font-size: 14px;
  }
  
  .cmd-button:hover {
    background: #2a2a2a;
    border-color: #666;
  }
  
  .cmd-button:active {
    background: #1e1e1e;
  }
  
  .host-display {
    margin-left: auto;
    background: #2a2a2a;
    color: #d4d4d4;
    padding: 8px 16px;
    border-radius: 4px;
    font-size: 13px;
    border: 1px solid #4d4d4d;
  }
  
  .content-container {
    flex: 1;
    overflow: hidden;
    position: relative;
  }
  
  .editor-container, .terminal-container {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    overflow: hidden;
    padding: 0 8px 8px 0;
    background: #1e1e1e;
    text-align: left;
  }
  
  .hidden {
    display: none;
  }
  
  .terminal-container ::-webkit-scrollbar {
    width: 10px;
  }
  
  .terminal-container ::-webkit-scrollbar-track {
    background: #1e1e1e;
  }
  
  .terminal-container ::-webkit-scrollbar-thumb {
    background: #666;
    border-radius: 5px;
  }
  
  .terminal-container ::-webkit-scrollbar-thumb:hover {
    background: #888;
  }
  </style>