# 使用

- 部署前：预下载
- 部署: 备份旧二进制文件，部署新二进制文件
- 部署后：检查，如果检查失败，则回滚（重新部署之前的版本）
- 回滚：如果有旧版本就执行回滚，如果没有就删除这次新部署的结果

## CD 下发

每个项目的 CD（版本发布） 都是这四个过程

## 服务的二进制和部署分离

版本发布（静态部分）：
● 部署和回滚 playbook 直接从 git 拉取：semaphore 负责
● 部署的环境变量直接从 git 拉取：semaphore 负责
● 部署的依赖包直接包含在 playbook 中： semaphore 负责

版本发布（服务交互）：

- runner-proxy 需要从 manager（版本发布控制台）领版本任务
- runner-proxy 下发任务并拆解到 semaphore
- semaphore 负责具体任务的执行
- runner-proxy 监控 semaphore 执行状态
- runner-proxy 收集任务执行结果
- 如果失败，则回滚：如果有旧版本就执行回滚，如果没有就删除这次新部署的结果

版本发布（必要信息）：runner-proxy 必须知道的信息（维护在 git 项目中）
    # 可配置参数 - 根据实际需求修改
    bin_tar: "streamd-20251026-040135.tar.gz" # 软件包名称
    depoloy_tar: "deploy_bin1.tar.gz" # 安装该软件包的代码包名称
    bin_tar_url: "<https://example.com/path/to/package>" # 下载URL
    depoloy_tar_url: "<https://example.com/path/to/package>" # 下载URL
    download_dir: "/tmp/download" # 下载保存目录
    bin_tar_dest: "/tmp/download/bin1.tar.gz" # 本地保存路径
    depoloy_bin_tar_dest: "/tmp/{{ software_name }}.tar.gz" # 本地保存路径
    # 解压相关参数（当 install_method 为 tarball 时使用）
    bin_extract_dir: "/tmp/extract/bin1"
    bin_extract_full_path: "{{ bin_extract_dir }}/bin1" # 二进制文件路径， 待部署是覆盖到服务真正运行的目录
    depoloy_bin_extract_dir: "/tmp/extract/bin1/deploy" # ansible helm terraform 等部署工具存放目录
    # install_method: "bin" # bin, container
