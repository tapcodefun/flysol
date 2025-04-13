<template>
  <LoginPage v-if="!isLogin" @login-success="handleLoginSuccess" />
  <div v-else class="host-management">
    <div class="operation-bar" style="position: relative;">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
      </el-button>
      <el-button type="success" @click="loadHostList">
        <el-icon><RefreshRight /></el-icon>
      </el-button>
      <el-input v-model="inputValue" style="width: 200px; margin: 0 10px;" placeholder="请输入主机名称或IP地址" />
      <el-button type="info" @click="handleButtonClick(inputValue)">
        <el-icon><Search /></el-icon>
      </el-button>
      <el-select v-model="selectedCategory" placeholder="分类" style="width: 120px; margin-right: 10px" clearable @change="filterByCategory">
        <el-option
          v-for="item in categories"
          :key="item"
          :label="item"
          @click="handleSortClick(item)"
        />
      </el-select>
      <div style="position:absolute; left: 500px; top:3px; display: flex; align-items: center; gap: 8px;">
        <ul style="list-style: none; margin: 0; padding: 0; display: flex; gap: 8px; overflow-x: auto; white-space: nowrap;">
          <li v-for="ip in [...new Set(hostList.map(host => host.masterIP).filter(Boolean))]" :key="ip" 
              style="padding: 4px 12px; background: #ecf5ff; border: 1px solid #d9ecff; border-radius: 4px; font-size: 13px; color: #409eff; white-space: nowrap;">
            {{ ip }}
          </li>
        </ul>
      </div>
    </div>
    <div v-if="isShow" style="display: flex; justify-content: flex-start;">
      <el-button type="primary" @click="handleCloseConnection">关闭</el-button>
    </div>
    <el-table :data="hostList" border style="width: 100%">
      <el-table-column prop="name" label="主机名称" width="180" />
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column prop="ip" label="IP地址" width="150" />
      <el-table-column prop="blockId" label="区块ID" width="120"/>
      <el-table-column prop="cpu" label="CPU" width="100" />
      <el-table-column prop="pid" label="进程" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.pid==-2" @click="handleFetch(row)" style="cursor: pointer;">拉取</el-tag>
          <el-tag v-else-if="row.pid==0">未启动</el-tag>
          <el-tag v-else-if="row.pid==-1">未安装</el-tag>
          <el-tag v-else>{{row.pid}}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-button size="small" type="success" v-if="row.status==='online'" @click="handleClose(row)">在线</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offline'" @click="handleStart(row)">离线</el-button>
          <el-button size="small" type="success" v-if="row.status==='onling'">开启中</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offling'">关闭中</el-button>
          <el-button size="small" type="danger" v-if="row.status==='offline' && row.pid>0" @click="handleClose(row)">关闭</el-button>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="success" @click="handleConnect(row)">连接</el-button>
            <el-button size="small" type="success" @click="showUploadDialog(row)">上传</el-button>

            <!-- 上传弹窗 -->
            <el-dialog v-model="uploadDialogVisible" title="上传" width="700px" top="10vh" append-to-body modal-append-to-body :modal="false" :show-close="false">
              <el-tabs>
                <el-tab-pane label="文件上传">
                  <el-upload
                    class="upload-area"
                    drag
                    action="#"
                    multiple
                    :auto-upload="false"
                    :on-change="handleFileChange"
                    ref="uploadRef"
                  >
                    <div class="upload-content">
                      <el-icon class="upload-icon"><upload-filled /></el-icon>
                      <div class="upload-text">将文件拖到此处，或<em>点击上传</em></div>
                    </div>
                    <template #tip>
                      <div class="el-upload__tip">已选择 {{ fileList.length }} 个文件</div>
                    </template>
                  </el-upload>
                  
                  <!-- SSH连接参数表单 -->
                  <div class="ssh-form" style="margin-top: 20px;">
                    <el-form :model="sshForm" label-width="100px">
                      <el-row :gutter="20">
                        <el-col :span="12">
                          <el-form-item label="目标IP" prop="targetIP">
                            <el-input v-model="sshForm.targetIP" placeholder="请输入目标主机IP" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="SSH用户" prop="sshUser">
                            <el-input v-model="sshForm.sshUser" placeholder="请输入SSH用户名" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="SSH密码" prop="sshPassword">
                            <el-input v-model="sshForm.sshPassword" type="password" show-password placeholder="请输入SSH密码" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="SSH端口" prop="sshPort">
                            <el-input v-model="sshForm.sshPort" placeholder="默认22" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="远程目录" prop="remoteDir">
                            <div style="display: flex; align-items: center;">
                              <el-input v-model="sshForm.remoteDir" placeholder="默认/home/bot" style="margin-right: 10px" />
                              <el-button type="primary" size="default" v-if="hasCompressedFile" @click="UPzip">解压</el-button>
                            </div>
                          </el-form-item>
                        </el-col>
                      </el-row>
                    </el-form>
                    <div class="dialog-footer">
                       <el-button @click="emptyFileList">取消</el-button>
                       <el-button type="primary" @click="handleUpload">确定</el-button>
                     </div>
                  </div>
                </el-tab-pane> 
                <el-tab-pane label="文件夹上传">
                  <div class="folder-upload-container">
                    <div class="upload-area folder-drop-area" @click="selectFolder" @dragover.prevent @drop.prevent="handleFolderDrop">
                      <div class="upload-content">
                        <el-icon class="upload-icon"><folder-add /></el-icon>
                        <div class="upload-text">将文件夹拖到此处，或<em>点击选择文件夹</em></div>
                      </div>
                    </div>
                    <div class="folder-preview-area" v-if="selectedFolderFiles.length > 0">
                      <div class="folder-info">
                        <div class="folder-item">
                          <el-icon class="folder-icon"><folder /></el-icon>
                          <span class="folder-name">{{ selectedFolderName }}</span>
                          <span class="file-count">({{ selectedFolderFiles.length }} 个文件)</span>
                          <el-button size="small" type="danger" class="delete-button" @click="clearSelectedFolder">删除</el-button>
                        </div>
                      </div>
                    </div>
                  </div>
                  <!-- SSH连接参数表单 -->
                  <div class="ssh-form" style="margin-top: 20px;">
                    <el-form :model="sshForm" label-width="100px">
                      <el-row :gutter="20">
                        <el-col :span="12">
                          <el-form-item label="目标IP" prop="targetIP">
                            <el-input v-model="sshForm.targetIP" placeholder="请输入目标主机IP" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="SSH用户" prop="sshUser">
                            <el-input v-model="sshForm.sshUser" placeholder="请输入SSH用户名" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="SSH密码" prop="sshPassword">
                            <el-input v-model="sshForm.sshPassword" type="password" show-password placeholder="请输入SSH密码" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="SSH端口" prop="sshPort">
                            <el-input v-model="sshForm.sshPort" placeholder="默认22" />
                          </el-form-item>
                        </el-col>
                        <el-col :span="12">
                          <el-form-item label="远程目录" prop="remoteDir">
                            <el-input v-model="sshForm.remoteDir" placeholder="默认/home/bot" />
                          </el-form-item>
                        </el-col>
                      </el-row>
                    </el-form>
                    <div class="dialog-footer">
                       <el-button @click="emptyFileList">取消</el-button>
                       <el-button type="primary" @click="handleUploadFileList">确定</el-button>
                     </div>
                  </div>
                </el-tab-pane>
              </el-tabs>
            </el-dialog>
            <el-button size="small" type="info" @click="handlePrivateKey(row)">私钥</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 自定义分类弹窗 -->
    <el-dialog v-model="categoryDialogVisible" title="自定义分类" width="400px">
      <el-form>
        <el-form-item label="分类名称">
          <el-input v-model="newCategory" placeholder="请输入分类名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="categoryDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleAddCategory">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 自定义主节点IP弹窗 -->
    <el-dialog v-model="masterIPDialogVisible" title="自定义主节点IP" width="400px">
      <el-form>
        <el-form-item label="主节点IP">
          <el-input v-model="newMasterIP" placeholder="请输入主节点IP" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="masterIPDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleAddMasterIP">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑主机' : '添加主机'" width="600px" top="10vh">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="主机名称" prop="name">
              <el-input v-model="formData.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item label="主机IP" prop="ip">
              <el-input v-model="formData.ip" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item label="私钥连接" prop="sshtype">
              <el-switch v-model="formData.sshtype" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item label="端口" prop="port">
              <el-input v-model="formData.port" type="number" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item label="账号" prop="username">
              <el-input v-model="formData.username" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item :label="formData.sshtype?'私钥密码':'登录密码'" prop="password" >
              <el-input v-model="formData.password" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="24" v-if="formData.install == 'wait' || formData.install == 'doing'">
            <el-form-item label="私钥" prop="private_key">
              <el-input type="textarea" row="4" v-model="formData.private_key" />
            </el-form-item>
          </el-col>
          <el-col  :span="24">
            <el-form-item label="描述" prop="desc">
              <el-input v-model="formData.desc" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="分类" prop="category">
              <div style="display: flex; align-items: center;">
                <el-select v-model="formData.category" style="width: 200px; margin-right: 10px;">
                  <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
                </el-select>
                <el-button type="primary" @click="categoryDialogVisible = true">添加分类</el-button>
                <el-button type="danger" @click="handleDeleteCategory(formData.category)">删除分类</el-button>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="主节点IP" prop="masterIP">
              <div style="display: flex; align-items: center;">
                <el-select v-model="formData.masterIP" style="width: 200px; margin-right: 10px;">
                  <el-option v-for="ip in masterIPs" :key="ip" :label="ip" :value="ip" />
                </el-select>
                <el-button type="primary" @click="masterIPDialogVisible = true">添加主节点IP</el-button>
              </div>
            </el-form-item>
          </el-col>
          <el-col  :span="24" v-if="formData.token">
            <el-form-item label="安全码" prop="token">
              <el-input v-model="formData.token" />
            </el-form-item>
          </el-col>
          <el-col :span="24" v-if="formData.install == 'finish'">
            <el-form-item label="钱包私钥" prop="siyao">
              <div style="margin-top: 5px;"> 
                <el-input v-model="siyao" type="password" style="float: left;width: 180px;margin-right: 5px;" /> 
                <el-button type="info" v-if="dosiyao==false" style="float: left;width: 60px;" @click="setSiyao(formData.ip,formData.password||'',formData.username,formData.port+'',formData.sshtype==true?formData.private_key:'')">更新</el-button>
                <el-button type="info" v-if="dosiyao==true" style="float: left;width: 60px;">更新中</el-button>
              </div>
            </el-form-item>
          </el-col>
          
          <el-col :span="24" v-if="formData.token">
            <el-form-item label="IP" prop="ips">
              <template v-if="typeof formData.ips != 'string'">
              <div style="margin-top: 5px;" v-for="(item,index) in formData.ips"> 
                <el-input v-model="item.ip" style="float: left;width: 160px;margin-right: 5px;" /> 
                <el-input v-model="item.iface" style="float: left;width: 80px;margin-right: 5px;" />
              </div>
              </template>
              <div style="margin-top: 5px;"> 
                <el-input v-model="newip.ip" style="float: left;width: 160px;margin-right: 5px;" /> 
                <el-input v-model="newip.iface" style="float: left;width: 80px;margin-right: 5px;" /> 
                <el-button type="info" style="float: left;width: 60px;" @click="addIP(formData.ip,formData.password||'',formData.username,formData.port+'',formData.sshtype==true?formData.private_key:'')">{{newip.status=='await'?"增加":"保存中"}}</el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div>
            <el-button type="info" @click="handleTestConnection" v-if="formData.install == 'wait'" >安装插件</el-button>
            <el-button type="info" v-if="formData.install == 'doing'" >安装中</el-button>
            <el-button type="info" v-if="formData.install == 'uninstall'">卸载中</el-button>
            <el-button type="info" @click="handleUninstall"  v-if="formData.install == 'finish'" >卸载</el-button>
            <el-button type="info" @click="setdefault"  v-if="formData.install == 'finish'" >重置</el-button>
            <el-button type="danger" @click="handleDelete(currentEditId)" v-if="isEdit">删除</el-button>
          </div>
          <div>
            <el-button @click="dialogVisible = false">取消</el-button>
            <el-button type="primary" @click="submitForm">确认</el-button>
          </div>
        </div>
      </template>
    </el-dialog>


  </div>
</template>



<script setup lang="ts">
import axios from 'axios';
import { ElLoading } from 'element-plus';
import { isMacOS,isWindows,OpenURL,loadEnvironment } from './utils/platform';
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox, install } from 'element-plus'
import { Plus, RefreshRight, Search } from '@element-plus/icons-vue'
import { UploadFileToRemoteHost,UploadPrivatekey,CreateSSHClient,OpenNewWindow,Install,Uninstall,RunServer,CloseServer,CheckPort,AddServerIP,Setprivatekey,Fetchost,UploadFolderToRemoteHost } from '../wailsjs/go/main/App'
import LoginPage from './components/LoginPage.vue'
import { is } from '@babel/types';

interface ipface {
  ip: string
  iface: string
  status: string
}
interface Host {
  id?: number
  name: string
  ip: string
  port: number
  username: string
  password?: string
  desc: string
  ips: ipface[] | ""
  rpc: string
  pid:number,
  grpc: string
  cpu: string
  token: string
  sshtype:boolean
  private_key:string
  masterIP: string
  status: 'online' | 'offline' | 'offing' | 'oning'
  install: 'doing' | 'finish' | 'wait' | 'uninstall'
  log:string
  category: string
  blockId: string
}

const formRef = ref<FormInstance>()
const hostList = ref<Host[]>([])
const originalHostList = ref<Host[]>([])
const dialogVisible = ref(false)
const uploadDialogVisible = ref(false)
const isEdit = ref(false)
const siyao = ref("")
const isLogin = ref(false)

const handleLoginSuccess = () => {
  isLogin.value = true
}
const dosiyao = ref(false)
const currentEditId = ref<number>(0)
const isShow = ref(false);
const selectedFolderName = ref('')
const selectedFolderFiles = ref<File[]>([])
const folderInput = ref<HTMLInputElement | null>(null)

const handleCloseConnection = () => {
  const connectionContainer = document.querySelector('div[style*="margin-top: 20px"]');
  if (connectionContainer) {
    connectionContainer.remove();
  }
  const hostManagement = document.querySelector('.host-management .el-table');
  if (hostManagement) {
    if (hostManagement instanceof HTMLElement) {
      hostManagement.style.display = 'block';
    }
  }
  const operationBar = document.querySelector('.host-management .operation-bar');
  if (operationBar) {
    if (operationBar instanceof HTMLElement) {
      operationBar.style.display = 'block';
    }
  }
  isShow.value = false;
};

const newip = ref<ipface>({iface:'',ip:'',status:'await'})

// 分类相关数据
const categoryDialogVisible = ref(false)
const newCategory = ref('')
const categories = ref<string[]>([])

// 主节点IP相关数据
const masterIPDialogVisible = ref(false)
const newMasterIP = ref('')
const masterIPs = ref<string[]>([])

const formData = reactive<Omit<Host, 'id' | 'status'>>({
  name: '',
  ip: '',
  port: 22,
  username: 'root',
  password: '',
  desc: '',
  ips: '',
  rpc: '',
  token:'',
  grpc: "",
  cpu: "",
  pid:-2,
  private_key:"",
  sshtype:false,
  install: "wait",
  log:'',
  category: '',
  masterIP: '',
  blockId: ''
})

const formRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [
    { required: true, message: '请输入IP地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入正确的IP地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { validator: (rule, value, callback) => {
        if (!value) {
          callback(new Error('请输入端口号'));
        } else if (!Number.isInteger(Number(value))) {
          callback(new Error('端口号必须为数字'));
        } else {
          callback();
        }
      }, trigger: 'blur' }
  ],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  masterIP: [{ required: true, message: '请选择主节点IP', trigger: 'change' }]
})

const initDB = (): Promise<IDBDatabase> => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('HostDB', 1)

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result
      if (!db.objectStoreNames.contains('hosts')) {
        const store = db.createObjectStore('hosts', { keyPath: 'id', autoIncrement: true })
        store.createIndex('ip', 'ip', { unique: false })
      }
    }

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

const getAllHosts = async (): Promise<Host[]> => {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('hosts', 'readonly')
    const store = transaction.objectStore('hosts')
    const request = store.getAll()

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

const saveHost = async (host: Host): Promise<void> => {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('hosts', 'readwrite')
    const store = transaction.objectStore('hosts')
    host.ips = ""
    const request = host.id ? store.put(host) : store.add(host)

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}
const deleteHost = async (id: number): Promise<void> => {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('hosts', 'readwrite')
    const store = transaction.objectStore('hosts')
    const request = store.delete(id)

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}


const loadHostList = async () => {
  try {
    const hosts = await getAllHosts();
    const updatedHosts = await Promise.all(hosts.map(async (host) => {
      try {
        const response = await axios.get(`http://${host.ip}:5189/metrics`, {
          headers: {'Content-Type': 'application/json'},
        });
        const content = response.data;
        host.status = 'online';
        host.token = content.token;
        host.cpu = (content.cpu).toFixed(2) + '%';
        host.ips = content.ips || [];
        const response2 = await axios.get(`http://${host.ip}:5189/pid`, {
          headers: {'Content-Type': 'application/json'},
        });
        host.pid = Number(response2.data.pid) || -2;
      } catch (err) {
        host.status = 'offline';
      }
      return host;
    }));
    hostList.value = updatedHosts;
    originalHostList.value = updatedHosts;
    // 更新分类列表
    categories.value = [...new Set(hostList.value.map((host: Host) => host.category))];
    // 更新主节点IP列表
    masterIPs.value = [...new Set(hostList.value.map((host: Host) => host.masterIP))];
  } catch (error) {
    console.error('加载数据失败:', error);
    ElMessage.error('数据加载失败');
  }
};

const newWindows = () =>{
  ElMessage.success('功能开发中');
  OpenNewWindow().then(res=>{
    console.log(res);
  })
}
const handleAddCategory = () => {
  if (!newCategory.value) {
    ElMessage.warning('请输入分类名称')
    return
  }
  if (categories.value.includes(newCategory.value)) {
    ElMessage.warning('该分类已存在')
    return
  }
  categories.value.push(newCategory.value)
  newCategory.value = ''
  categoryDialogVisible.value = false
  ElMessage.success('添加成功')
}

const handleAddMasterIP = () => {
  if (!newMasterIP.value) {
    ElMessage.warning('请输入主节点IP')
    return
  }
  if (masterIPs.value.includes(newMasterIP.value)) {
    ElMessage.warning('该主节点IP已存在')
    return
  }
  if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(newMasterIP.value)) {
    ElMessage.warning('请输入正确的IP地址格式')
    return
  }
  masterIPs.value.push(newMasterIP.value)
  newMasterIP.value = ''
  masterIPDialogVisible.value = false
  ElMessage.success('添加成功')
}

const handleAdd = () => {
  isEdit.value = false
  dialogVisible.value = true
  Object.assign(formData, {
    name: '127.0.0.1',
    ip: '127.0.0.1',
    port: 22,
    username: 'root',
    password: 'root',
    desc: '',
    install:"wait"
  })
}
const setdefault = async () => {
  if (!formRef.value) return;
  try {
    // 修复：如果是新增操作，移除 id 字段
    const hostData: Host = {
      ...formData,
      status: 'offline'
    };
    formData.install = "wait";
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
};

const handleEdit = (row: Host) => {
  isEdit.value = true
  currentEditId.value = row.id!
  if(row.ips.length>0){
    var ip = row.ips[0];
    if(typeof ip === 'string'){
      newip.value.iface = "";
    }else{
      newip.value.iface = ip.iface;
    }
  }else{
    newip.value.iface = "";
  }
  newip.value.ip = "";
  newip.value.status = "await";
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = async (id: number) => {
  ElMessageBox.confirm('确认删除该主机？', '警告', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteHost(id)
      const index = hostList.value.findIndex(host => host.id === id);
      hostList.value.splice(index, 1);
      await loadHostList()
      dialogVisible.value = false
      ElMessage.success('删除成功')
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// const handleUp = async (host: Host) => {
//   const index = hostList.value.findIndex(h => h.id === host.id);
//   hostList.value[index].status = 'online';
// }

const handleClose = async (host: Host) => {
  ElMessage.info(`正在关闭 ${host.ip}:${host.port}`);

  // 找到当前主机在列表中的索引
  const index = hostList.value.findIndex(h => h.id === host.id);
  if (index === -1) return; // 如果找不到主机，直接返回

  // 更新主机状态为 "关闭中"
  hostList.value[index].pid = -2;
  hostList.value[index].status = 'offing';

  try {
    // 调用关闭服务器的函数
    CloseServer(host.ip, host.password || '',host.username, host.port+"",host.sshtype==true?host.private_key:"");

    let attemptCount = 0; // 尝试次数计数器
    const maxAttempts = 10; // 最大尝试次数
    const pollInterval = 1000; // 轮询间隔时间（1秒）

    // 定义轮询函数
    const pollServerStatus = async () => {
      try {
        // 发送请求检查服务器状态
        const response = await axios.get(`http://${host.ip}:5189/metrics`, {
          headers: {'Content-Type': 'application/json'},
        });

        // 如果服务器仍在运行，增加尝试次数
        attemptCount++;
        if (attemptCount >= maxAttempts) {
          // 达到最大尝试次数，停止轮询并提示关闭失败
          ElMessage.error('关闭失败');
          hostList.value[index].status = 'offline'; // 更新状态为离线
          return;
        }

        // 继续轮询
        setTimeout(pollServerStatus, pollInterval);
      } catch (err) {
        // 如果请求失败，说明服务器已关闭
        ElMessage.success('关闭成功');
        hostList.value[index].status = 'offline'; // 更新状态为离线
      }
    };

    // 开始轮询
    pollServerStatus();
  } catch (error) {
    // 处理关闭过程中的异常
    console.error('关闭失败:', error);
    ElMessage.error('关闭异常');
    hostList.value[index].status = 'offline'; // 更新状态为离线
  }
};

const handleStart = async (host: Host) => {
  ElMessage.info(`正在连接 ${host.ip}:${host.port}`);

  // 找到当前主机在列表中的索引
  const index = hostList.value.findIndex(h => h.id === host.id);
  if (index === -1) return; // 如果找不到主机，直接返回

  // 更新主机状态为 "启动中"
  hostList.value[index].status = 'oning';

  try {
    // 生成随机 token
    const token = generateRandomString(16);
    hostList.value[index].token = token;
    console.log("sshtype",host.sshtype);
    // 启动服务器
    const res = await RunServer(host.ip,token, host.password || '', host.username, host.port+"",host.sshtype==true?host.private_key:"");
    console.log(res);
    hostList.value[index].log = res;

    let attemptCount = 0; // 尝试次数计数器
    const maxAttempts = 10; // 最大尝试次数
    const pollInterval = 1000; // 轮询间隔时间（1秒）

    // 定义轮询函数
    const pollServerStatus = async () => {
      try {
        // 检查端口状态
        const res = await CheckPort(host.ip, host.password || '',host.username, host.port+"");
        console.log(res);

        if (res === 'success') {
          // 如果启动成功，停止轮询并更新状态
          ElMessage.success('启动成功');
          hostList.value[index].status = 'online';
          // 获取区块ID
          if (hostList.value[index].masterIP) {
            try {
              const blockResponse = await axios.get(`https://sol.tapcode.fun/api/solana/slot?ip=${hostList.value[index].masterIP}`);
              hostList.value[index].blockId = blockResponse.data.data.slot;
            } catch (err) {
              hostList.value[index].blockId = '-';
            }
          }
          return;
        } else {
          // 如果未成功，增加尝试次数
          attemptCount++;
          if (attemptCount >= maxAttempts) {
            // 达到最大尝试次数，停止轮询并提示启动失败
            ElMessage.error('启动失败');
            hostList.value[index].status = 'offline'; // 更新状态为离线
            return;
          }

          // 继续轮询
          setTimeout(pollServerStatus, pollInterval);
        }
      } catch (err) {
        // 如果检查端口状态失败，停止轮询并提示启动异常
        console.error('检查端口状态失败:', err);
        ElMessage.error('启动异常');
        hostList.value[index].status = 'offline'; // 更新状态为离线
      }
    };

    // 开始轮询
    pollServerStatus();
  } catch (error) {
    // 处理启动过程中的异常
    console.error('启动失败:', error);
    ElMessage.error('启动异常');
    hostList.value[index].status = 'offline'; // 更新状态为离线
  }
};

const handleFetch = async (row: Host) => {
  const index = hostList.value.findIndex(h => h.id === row.id);
  try {
    Fetchost(row.ip+":5189").then(async (res: any) => {
      var content = JSON.parse(res)
      console.log(content)
      hostList.value[index].token = content.token;
      hostList.value[index].cpu = (content.cpu).toFixed(2) + '%';
      hostList.value[index].pid = Number(content.pid) || -2;
      if(content.pid>0){
        hostList.value[index].status = 'online';
        // 更新区块ID
        if (hostList.value[index].masterIP) {
          try {
            const blockResponse = await axios.get(`https://sol.tapcode.fun/api/solana/slot?ip=${hostList.value[index].masterIP}`);
            hostList.value[index].blockId = blockResponse.data.data.slot;
          } catch (err) {
            hostList.value[index].blockId = '-';
          }
        }
      }else{
        hostList.value[index].status = 'offline';
        hostList.value[index].blockId = '-';
      }
    });
  } catch (error) {
    ElMessage.error('获取失败');
  }
}
const handleConnect = (host: Host) => {
  const index = hostList.value.findIndex(h => h.id === host.id);
  var url = 'http://'+host.ip+':5189?model=run&code='+host.token;
  axios.get("http://"+host.ip+":5189/metrics", {
    headers: {'Content-Type': 'application/json'}
  })
  .then(() => {
    host.status = 'online';
    hostList.value[index] = host;
    // 隐藏host-management元素
    const hostManagement = document.querySelector('.host-management .el-table');
    if (hostManagement) {
      if (hostManagement instanceof HTMLElement) {
        hostManagement.style.display = 'none';
      }
    }
    const operationBar = document.querySelector('.host-management .operation-bar');
    if (operationBar) {
      if (operationBar instanceof HTMLElement) {
        operationBar.style.display = 'none';
      }
    }
    // 在页面底部添加连接显示区域
    isShow.value = !isShow.value;
    const connectionContainer = document.createElement('div');
    connectionContainer.style.marginTop = '20px';
    connectionContainer.style.height = '100vh';
    connectionContainer.style.width = '100%';
    connectionContainer.style.backgroundColor = '#1e1e1e';
    connectionContainer.style.color = '#ffffff';
    connectionContainer.innerHTML = `
      <h3>连接到 ${host.name}</h3>
      <iframe src="${url}" style="width: 100%; height: 100%; border: none;"></iframe>
    `;
    const table = document.querySelector('.el-table');
    table?.parentNode?.insertBefore(connectionContainer, table.nextSibling);
  }).catch(async()=>{
    ElMessage.error('连接失败');
  })
}

const handleUninstall = async () => {
  if (!formRef.value) return;
  try {
    // 修复：如果是新增操作，移除 id 字段
    const hostData: Host = {
      ...formData,
      status: 'offline'
    };
    formData.install = "uninstall";
    if(hostData.password==''){
      return ElMessage.error('密码不能为空');
    }
    ElMessage.info(`正在卸载 ${hostData.ip}`);
    Uninstall(hostData.ip, hostData.password || '',hostData.username,hostData.port+"",hostData.sshtype==true?hostData.private_key:"").then((res) => {
      console.log('Uninstall',res)
      if (res === 'success') {
        formData.install = "wait";
        ElMessage.success('安装成功');
      } else {
        formData.install = "finish";
        ElMessage.error('安装失败' + res);
      }
    });
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
}
const handleTestConnection = async () => {
  if (!formRef.value) return;
  try {
    // 修复：如果是新增操作，移除 id 字段
    const hostData: Host = {
      ...formData,
      status: 'offline'
    };
    // 使用 Vue 的响应式方法确保更新
    formData.install = "doing";

    if(hostData.password==''){
      formData.install = "wait";
      return ElMessage.error('密码不能为空');
    }
    ElMessage.info(`正在连接 ${hostData.ip}`);
    Install(hostData.ip,hostData.private_key, hostData.password || '',hostData.username,hostData.port+"").then((res) => {
      if (res === 'success') {
        // 使用 Vue 的响应式方法确保更新
        formData.install = "finish";
        ElMessage.success('安装成功');
      } else {
        // 使用 Vue 的响应式方法确保更新
        formData.install = "wait";
        ElMessage.error('安装失败' + res);
      }
    });
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
};
const setSiyao = async (ip:string, password:string,username:string,port:string,private_key:string) => {
  try {
    dosiyao.value = true
    Setprivatekey(ip,password,siyao.value,username,port+"",private_key).then((res:any) => {
      dosiyao.value = false
      if (res === 'success') {
        siyao.value = ""
        ElMessage.success('设置成功');
      } else {
        ElMessage.error('设置失败' + res);
      }
    });
  } catch (error) {
    dosiyao.value = false
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
}

const addIP = async (ip:string, password:string,username:string,port:string,private_key:string) => {
  if(newip.value.ip == ''){
    return ElMessage.error('IP不能为空');
  }
  if(newip.value.status != 'await'){
    return ElMessage.error("请等待上次操作完成");
  }
  try {
    newip.value.status = 'doing'
    AddServerIP(ip,password,newip.value.ip,newip.value.iface,username,port,private_key).then((res:any) => {
      newip.value.status = 'await'
      if (res === 'success') {
        ElMessage.success('添加成功');
      } else {
        ElMessage.error('添加失败' + res);
      }
    });
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
};
const submitForm = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    const hostData: Host = {
      ...formData,
      status: 'online',
      install: "finish"
    };
    hostData.ip = hostData.ip.trim();
    hostData.password = hostData.password?.trim();
    if (isEdit.value && currentEditId.value) {
      // 编辑操作时，保留 id
      hostData.id = currentEditId.value;
    }else{
      const index = hostList.value.findIndex(h => h.ip === hostData.ip);
      if (index > -1) {
        return ElMessage.error('不能重复添加');
      }
      delete hostData.id
    }
    await saveHost(hostData);
    await loadHostList();

    dialogVisible.value = false;
    ElMessage.success(isEdit.value ? '修改成功' : '添加成功');
  } catch (error) {
    console.error('操作失败:', error);
    ElMessage.error('操作失败');
  }
};

const generateRandomString = (length: number) => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let randomString = '';

    for (let i = 0; i < length; i++) {
        const randomIndex = Math.floor(Math.random() * chars.length);
        randomString += chars[randomIndex];
    }

    return randomString;
}


const inputValue = ref('')
const selectedCategory = ref('')

const filterByCategory = () => {
  if (!selectedCategory.value) {
    hostList.value = originalHostList.value
    return
  }
  hostList.value = originalHostList.value.filter(host => host.category === selectedCategory.value)
}
// 分类功能
const handleSortClick = (value: string) => {
  hostList.value = originalHostList.value;
  hostList.value = hostList.value.filter((host) => host.category === value)
}

//搜索功能
const handleButtonClick = (value: string) => {
  hostList.value = originalHostList.value;
  hostList.value = hostList.value.filter((host) => host.name.includes(value) || host.ip.includes(value))
}

const showUploadDialog = (row: Host) => {
  // 设置默认值
  sshForm.value.targetIP = row.ip;
  sshForm.value.sshUser = 'root'; // 默认用户名设为root
  sshForm.value.sshPassword = row.password || ''; // 如果有密码则使用，否则为空
  sshForm.value.sshPort = row.port ? row.port.toString() : '22';
  uploadDialogVisible.value = true;
}

const fileList = ref<File[]>([]);
const uploadRef = ref();
const hasCompressedFile = ref(false);
const isUpzip = ref(false);

// 检查是否为压缩文件
const isCompressedFile = (fileName: string): boolean => {
  const compressedExtensions = [".zip", ".tar", ".gz", ".rar", ".7z"];
  return compressedExtensions.some(ext => fileName.toLowerCase().endsWith(ext));
};

const UPzip = () => {
 isUpzip.value = true 
 ElMessage.info('文件在上传后将被解压')
}

// SSH连接参数表单数据
const sshForm = ref({
  targetIP: '',
  sshUser: '',
  sshPassword: '',
  sshPort: '22',
  remoteDir: '/home/bot'
});

const handleFileChange = (file: File, files: FileList) => {
  fileList.value = Array.from(files);
  // 检查是否包含压缩文件
  hasCompressedFile.value = fileList.value.some(file => isCompressedFile(file.name));
};

const emptyFileList = () => {
  uploadDialogVisible.value = false;
  fileList.value = [];
  hasCompressedFile.value = false;
  selectedFolderName.value = ''
  selectedFolderFiles.value = []
  if (uploadRef.value) {
    uploadRef.value.clearFiles();
  }
};

// 处理文件上传
const handleUpload = async () => {
  if (fileList.value.length === 0) {
    ElMessage.warning('请选择要上传的文件');
    return;
  }
  
  // 验证SSH连接参数
  if (!sshForm.value.targetIP || !sshForm.value.sshUser || !sshForm.value.sshPassword) {
    ElMessage.warning('请填写完整的SSH连接信息');
    return;
  }

  // 验证远程目录路径格式
  if (!sshForm.value.remoteDir.startsWith('/')) {
    ElMessage.warning('远程目录路径必须以"/"开头');
    return;
  }
  if (sshForm.value.remoteDir.includes('\\')) {
    ElMessage.warning('远程目录路径不能包含"\\"字符');
    return;
  }
  
  const loading = ElLoading.service({
    lock: true,
    text: '正在上传文件，请稍候...',
    background: 'rgba(0, 0, 0, 0.7)'
  });
  
  try {
    let successCount = 0;
    let failCount = 0;
    
    // 为每个文件创建FormData对象
    for (const file of fileList.value) {
      const formData = new FormData();
      formData.append('file', file);
      
      try {
        // 读取文件内容为Base64
        const fileContent = await readFileAsBase64((file as any).raw);
        
        // 调用后端API上传文件
        const result = await UploadFileToRemoteHost(
          sshForm.value.targetIP,
          sshForm.value.sshUser,
          sshForm.value.sshPassword,
          sshForm.value.sshPort,
          sshForm.value.remoteDir,
          file.name,
          fileContent,
          String(isUpzip.value)
        );

        isUpzip.value = false

        if (result === 'success') {
          successCount++;
          ElMessage.success(`文件 ${file.name} 上传成功`);
        } else {
          failCount++;
          ElMessage.error(`文件 ${file.name} 上传失败: ${result}`);
        }
      } catch (error) {
        failCount++;
        console.error(`文件 ${file.name} 上传出错:`, error);
        ElMessage.error(`文件 ${file.name} 上传失败`);
      }
    }
    
    // 显示总体上传结果
    loading.close();
    if (successCount > 0) {
      ElMessage.success(`成功上传 ${successCount} 个文件`);
    }
    if (failCount > 0) {
      ElMessage.warning(`${failCount} 个文件上传失败`);
    }
    
    // 关闭上传对话框
    uploadDialogVisible.value = false;
    // 清空文件列表
    fileList.value = [];
    if (uploadRef.value) {
      uploadRef.value.clearFiles();
    }
  } catch (error) {
    loading.close();
    console.error('文件上传失败:', error);
    ElMessage.error('文件上传失败: ' + error);
  }
};

// 将文件读取为Base64编码
const readFileAsBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      // 获取Base64编码，去掉前缀
      const base64 = reader.result as string;
      const base64Content = base64.split(',')[1];
      resolve(base64Content);
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
};


//获取私钥文件
const handlePrivateKey = async (row: any) => {
  try {
    ElMessage.info('正在连接并获取私钥文件...')
    const privateKey = await UploadPrivatekey(row.ip, row.username, row.password, row.port+'')
    if (!privateKey.includes('失败')) {
      ElMessage.success('私钥文件已下载到桌面')
    } else {
      ElMessage.error(privateKey)
    }
  } catch (error) {
    ElMessage.error('操作失败：' + error)
  }
}

// 文件夹上传
const selectFolder = () => {
  // 创建input元素用于选择文件夹
  if (!folderInput.value) {
    const input = document.createElement('input')
    input.type = 'file'
    input.webkitdirectory = true
    input.style.display = 'none'
    input.onchange = (event: Event) => {
      const files = (event.target as HTMLInputElement).files
      if (files && files.length > 0) {
        // 获取文件夹名称（从第一个文件的路径中提取）
        const firstFile = files[0]
        const folderPath = firstFile.webkitRelativePath
        selectedFolderName.value = folderPath.split('/')[0]
        
        // 更新已选择的文件列表
        selectedFolderFiles.value = Array.from(files)
      }
      // 重置input以允许选择相同的文件夹
      input.value = ''
    }
    document.body.appendChild(input)
    folderInput.value = input
  }
  folderInput.value.click()
}

const handleFolderDrop = (event: DragEvent) => {
  const items = event.dataTransfer?.items
  if (!items) return

  // 检查是否拖拽了文件夹
  const entry = items[0].webkitGetAsEntry()
  if (entry && entry.isDirectory) {
    const files: File[] = []
    const traverseDirectory = (entry: any, path: string) => {
      if (entry.isFile) {
        entry.file((file: File) => {
          Object.defineProperty(file, 'webkitRelativePath', {
            value: path + '/' + file.name
          })
          files.push(file)
        })
      } else if (entry.isDirectory) {
        const dirReader = entry.createReader()
        dirReader.readEntries((entries: any[]) => {
          for (const entry of entries) {
            traverseDirectory(entry, path ? path + '/' + entry.name : entry.name)
          }
        })
      }
    }
    
    traverseDirectory(entry, '')
    selectedFolderName.value = entry.name
    selectedFolderFiles.value = files
  }
}

const clearSelectedFolder = () => {
  selectedFolderName.value = ''
  selectedFolderFiles.value = []
}
const handleUploadFileList = async () => {
  if (selectedFolderFiles.value.length === 0) {
    ElMessage.warning('请选择要上传的文件夹');
    return;
  }
  
  // 验证SSH连接参数
  if (!sshForm.value.targetIP || !sshForm.value.sshUser || !sshForm.value.sshPassword) {
    ElMessage.warning('请填写完整的SSH连接信息');
    return;
  }

  // 验证远程目录路径格式
  if (!sshForm.value.remoteDir.startsWith('/')) {
    ElMessage.warning('远程目录路径必须以"/"开头');
    return;
  }
  if (sshForm.value.remoteDir.includes('\\')) {
    ElMessage.warning('远程目录路径不能包含"\\"字符');
    return;
  }
  
  ElMessage.info('正在准备上传文件夹，请稍候...');
  
  try {
    // 创建文件内容映射对象，键为相对路径，值为Base64编码的文件内容
    const folderContent: Record<string, string> = {};
    const folderName = selectedFolderName.value;
    
    // 收集所有文件的内容
    const loadingInstance = ElLoading.service({
      fullscreen: true,
      text: '正在读取文件内容...',
      background: 'rgba(0, 0, 0, 0.7)'
    });
    
    try {
      // 并行读取所有文件内容
      const fileReadPromises = selectedFolderFiles.value.map(async (file) => {
        try {
          const relativePath = file.webkitRelativePath;
          const fileContent = await readFileAsBase64(file);
          folderContent[relativePath] = fileContent;
        } catch (error) {
          console.error(`读取文件 ${file.webkitRelativePath} 失败:`, error);
          throw error;
        }
      });
      
      await Promise.all(fileReadPromises);
      loadingInstance.close();
      
      ElMessage.info(`正在上传文件夹 ${folderName}，共 ${selectedFolderFiles.value.length} 个文件...`);
      
      // 调用后端API一次性上传整个文件夹
      const result = await UploadFolderToRemoteHost(
        sshForm.value.targetIP,
        sshForm.value.sshUser,
        sshForm.value.sshPassword,
        sshForm.value.sshPort,
        sshForm.value.remoteDir,
        folderName,
        folderContent
      );
      
      if (result === 'success') {
        ElMessage.success(`文件夹 ${folderName} 上传成功，共 ${selectedFolderFiles.value.length} 个文件`);
      } else {
        ElMessage.error(`文件夹上传失败: ${result}`);
      }
      
      // 关闭上传对话框并清空文件列表
      uploadDialogVisible.value = false;
      clearSelectedFolder();
    } catch (error) {
      loadingInstance.close();
      throw error;
    }
  } catch (error) {
    console.error('文件夹上传失败:', error);
    ElMessage.error('文件夹上传失败: ' + error);
  }
}

const handleDeleteCategory = (category: string) => {
  if (!category) {
    ElMessage.warning('请先选择要删除的分类')
    return
  }
  
  const hasHostsInCategory = hostList.value.some(host => host.category === category)
  
  if (hasHostsInCategory) {
    ElMessage.warning('该分类下存在主机，无法删除')
    return
  }
  
  const index = categories.value.indexOf(category)
  if (index > -1) {
    categories.value.splice(index, 1)
    ElMessage.success('分类删除成功')
    formData.category = ''
  }
}

onMounted(async() => {
  await loadEnvironment()
  loadHostList()
})
</script>

<style scoped>
.host-management {
  padding: 20px;
}

.operation-bar {
  margin-bottom: 20px;
  float: left;
}

.el-table {
  margin-top: 20px;
}

.el-button + .el-button {
  margin-left: 8px;
}

.upload-demo {
  position: relative;
}

.upload-demo .el-upload__text {
  position: relative;
  z-index: 1;
  white-space: nowrap;
}

/* 上传对话框样式 */
.upload-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.upload-area {
  width: 100%;
  height: 300px;
  margin-bottom: 20px;
  position: relative;
}

.el-upload-list {
  max-height: 250px;
  overflow-y: auto;
  padding-right: 5px;
  scrollbar-width: thin;
}

.el-upload-list::-webkit-scrollbar {
  width: 6px;
}

.el-upload-list::-webkit-scrollbar-thumb {
  background-color: #909399;
  border-radius: 3px;
}

.el-upload-list::-webkit-scrollbar-track {
  background-color: #f5f7fa;
}

.upload-content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.upload-area:hover {
  border-color: #409eff;
}

.upload-icon {
  font-size: 48px;
  color: #c0c4cc;
  margin-bottom: 10px;
}

.upload-text {
  color: #606266;
  font-size: 14px;
}

.upload-buttons {
  display: flex;
  justify-content: center;
  gap: 10px;
}

/* 文件夹上传容器样式 */
.folder-upload-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 300px;
}

.folder-drop-area {
  height: 50%;
  width: 100%;
  min-height: 150px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.3s;
}

.folder-drop-area:hover {
  border-color: #409eff;
}

.folder-preview-area {
  height: auto;
  margin-top: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 10px;
}

.folder-info {
  display: flex;
  justify-content: center;
}

.folder-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  background-color: #f5f7fa;
  position: relative;
}

.folder-icon {
  font-size: 20px;
  color: #909399;
  margin-right: 8px;
}

.folder-name {
  font-weight: bold;
  color: #303133;
  margin-right: 10px;
}

.file-count {
  color: #909399;
  margin-right: 10px;
}

.delete-button {
  margin-left: auto;
}
</style>
