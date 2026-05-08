const API_BASE = 'https://video-saver.yulin.ca'

Page({
  data: {
    url: '',
    videoUrl: '',
    loading: false,
    downloading: false,
    progress: 0,
    saved: false,
    error: '',
  },

  onInput(e) {
    this.setData({ url: e.detail.value, error: '' })
  },

  onResolve() {
    const url = this.data.url.trim()
    if (!url) {
      this.setData({ error: '请粘贴分享链接' })
      return
    }

    this.setData({ loading: true, error: '' })

    wx.request({
      url: `${API_BASE}/api/resolve`,
      method: 'POST',
      header: { 'Content-Type': 'application/json' },
      data: { url },
      success: (res) => {
        if (res.data.ok) {
          this.setData({ videoUrl: res.data.videoUrl })
        } else {
          this.setData({ error: res.data.error || '解析失败' })
        }
      },
      fail: () => this.setData({ error: '网络错误，请检查连接' }),
      complete: () => this.setData({ loading: false }),
    })
  },

  onDownload() {
    this.setData({ downloading: true, progress: 0 })

    const task = wx.downloadFile({
      url: this.data.videoUrl,
      success: (res) => {
        if (res.statusCode !== 200) {
          wx.showToast({ title: '下载失败', icon: 'error' })
          this.setData({ downloading: false })
          return
        }
        wx.saveVideoToPhotosAlbum({
          filePath: res.tempFilePath,
          success: () => this.setData({ downloading: false, saved: true }),
          fail: (err) => {
            if (err.errMsg.includes('auth deny')) {
              wx.showModal({
                title: '需要相册权限',
                content: '请在设置中允许访问相册',
                confirmText: '去设置',
                success: (modal) => { if (modal.confirm) wx.openSetting() },
              })
            } else {
              wx.showToast({ title: '保存失败', icon: 'error' })
            }
            this.setData({ downloading: false })
          },
        })
      },
      fail: () => {
        wx.showToast({ title: '下载失败', icon: 'error' })
        this.setData({ downloading: false })
      },
    })

    task.onProgressUpdate((res) => {
      this.setData({ progress: res.progress })
    })
  },

  onReset() {
    this.setData({ url: '', videoUrl: '', progress: 0, saved: false, error: '' })
  },
})
