<script lang="ts" setup>
import Terminal from './terminal/index.vue';
import {Greet} from '../../wailsjs/go/main/App'
import XtermComponent from './XtermComponent.vue';
import TerminalPage from './TerminalPage.vue';
</script>

<template>
  <main>
    <!-- <Terminal
        :style="{
            height: `calc(100vh - 269px)`,
            'background-color': `var(--panel-logs-bg-color)`,
        }"
        :ref="'t-s'"
    ></Terminal> -->
      <!-- <XtermComponent></XtermComponent> -->
    <TerminalPage/>
  </main>
</template>

<style scoped>
.result {
  height: 20px;
  line-height: 20px;
  margin: 1.5rem auto;
}

.input-box .btn {
  width: 60px;
  height: 30px;
  line-height: 30px;
  border-radius: 3px;
  border: none;
  margin: 0 0 0 20px;
  padding: 0 8px;
  cursor: pointer;
}

.input-box .btn:hover {
  background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
  color: #333333;
}

.input-box .input {
  border: none;
  border-radius: 3px;
  outline: none;
  height: 30px;
  line-height: 30px;
  padding: 0 10px;
  background-color: rgba(240, 240, 240, 1);
  -webkit-font-smoothing: antialiased;
}

.input-box .input:hover {
  border: none;
  background-color: rgba(255, 255, 255, 1);
}

.input-box .input:focus {
  border: none;
  background-color: rgba(255, 255, 255, 1);
}
</style>
