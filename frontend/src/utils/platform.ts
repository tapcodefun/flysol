import { Environment,BrowserOpenURL } from '../../wailsjs/runtime'

let os = ''

export async function loadEnvironment() {
    const env = await Environment()
    os = env.platform
}

export function OpenURL(url:string) {
    return BrowserOpenURL(url)
}
export function isMacOS() {
    return os === 'darwin'
}

export function isWindows() {
    console.log(os)
    return os === 'windows'
}
