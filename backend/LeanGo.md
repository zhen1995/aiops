# go mod tidy和go get xxx区别
## go mod tidy 作用
+ 扫描项目中所有.go文件中的import
- 确保 go.mod 和实际代码完全匹配
- 常用于： 拉取新代码后同步依赖、删除代码后清理无用依赖、CI/CD 构建前确保依赖完整

## go get xxx
- 显示下载并添加某个包到go.mod
常用于：引入项目原本没有的新依赖、升级某个依赖到特定/最新的版本

## 最佳实践：
1.添加新包时：先用go get xx，再用go mod tidy 确保干净
2.日常开发/提交前：运行 go mod tidy 保证依赖整洁
3.不要手动编辑 go.mod :用这两个命令管理即可。
