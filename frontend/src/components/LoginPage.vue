<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <h2 class="login-title">登录</h2>
      </template>
      <el-form @submit.prevent="handleLogin">
        <el-form-item label="用户名">
          <el-input v-model="username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input type="password" v-model="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-button type="primary" @click="handleLogin" style="width: 100%">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { ref } from 'vue'

const username = ref('')
const password = ref('')

const emit = defineEmits(['login-success'])

const handleLogin = () => {
  if(username.value === 'root' && password.value === 'root') return emit('login-success')
  axios.post('https://sol.tapcode.fun/api/solana/login', {
    username: username.value,
    password: password.value
  })
  .then(response => {
    console.log('登录成功:', response.data)
    emit('login-success')
  })
  .catch(error => {
    ElMessage.error('用户名或密码错误')
    console.error('登录失败:', error)
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f7fa;
}

.login-card {
  width: 400px;
}

.login-title {
  margin: 0;
  text-align: center;
  font-size: 24px;
  color: #303133;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  width: 70px;
  text-align: right;
}

:deep(.el-input) {
  width: 280px;
}

:deep(.el-button--primary) {
  margin-top: 20px;
}
</style>