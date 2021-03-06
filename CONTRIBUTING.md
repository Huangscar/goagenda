# Contributing

更新于2018.10.18。

本文适用于《服务计算》课程同一小组的同学。

## Coding and Pull Request

1. 前往[Teambition](https://www.teambition.com/project/5bc6ffbaf10ae90018184bd0/)领取任务。
2. Fork本项目。
3. 克隆你的项目到本地。

```sh
$ git clone https://github.com/yourname/goagenda.git $GOPATH/src/github.com/MegaShow/goagenda
```

4. 编写任务逻辑，且自己Review。
5. Pull request至本项目。

### Coding Require

* Verbose日志可出现在任何函数中。
* Service服务函数中不应该出现任何Info、Show、Error日志，而是通过返回error的方式交给Controller打印日志。

### Commit Require

* Message Title只允许出现英文，首字母大写，禁止以标点符号结尾。
* Message Body只允许出现英文，允许留空，其余要求不做限制。

## Code Architecture

### 如何编写一个命令集？

> **Go Agenda的命令集架构已编写好，但阅读本节有助于了解Go Agenda的架构。**
>
>
>
> Go Agenda使用**Command、Controller、Services、Models**的架构形式实现逻辑。
>
> 本节将描述Command与Controller的关系。

什么是命令集？在Go Agenda中，我们将一系列类似、相同对象的命令的集合称为命令集。命令集方便我们对数量庞大的命令加以管理，我们将同一个命令集的命令定义在同一个`.go`文件中。比如，下列命令是一个命令集，并且下列命令均有相同的父命令。

```sh
$ agenda user set
$ agenda user list
$ agenda user delete
```

当然，不相似的命令一样可以组成命令集，并不要求均有相同的父命令。在Go Agenda中，以下的命令均在同一个`.go`文件中定义。

```sh
$ agenda register
$ agenda login
$ agenda logout
$ agenda status
```

为什么使用命令集？Go Agenda将同一个命令集的命令定义在同一个`.go`文件中，并且同一个命令集的命令共享一个Controller接口类型。

下面我们将以编写命令集`user`为例，介绍Go Agenda的Command、Controller前面两层模型是如何工作的。(为了简化代码，部分细节或属性省去)

首先，需要编写命令集的根命令(父命令)，如果命令集内的命令不具有相同父命令，那再按实际情况修改代码。

```go
var userRootCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"u"},
}
```

然后编写命令集三个命令。

```go
var userDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Run:     wrapper(controller.GetUserCtrl().UserDelete),
}

var userListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Run:     wrapper(controller.GetUserCtrl().UserList),
}

var userSetCmd = &cobra.Command{
	Use:     "set",
	Aliases: []string{"s"},
	Run:     wrapper(controller.GetUserCtrl().UserSet),
}
```

现在你可能疑惑`Run`属性中的`wrapper`函数有什么用，我们可以先不管它，先来编写导入命令和处理命令Argument、Flag的逻辑。

在`cmd/*.go`文件的`init`函数中，将父命令添加到`rootCmd`中，并将命令集的命令均添加到该父命令中。对于Argument和Flag的处理，均采用`XxxxP`的形式处理。

```go
func init() {
	rootCmd.AddCommand(userRootCmd)
	userRootCmd.AddCommand(userDeleteCmd)
	userRootCmd.AddCommand(userListCmd)
	userRootCmd.AddCommand(userSetCmd)

	userListCmd.Flags().StringP("user", "u", "", "the username searched")

	userSetCmd.Flags().StringP("password", "p", "", "new password")
	userSetCmd.Flags().StringP("email", "e", "", "new email")
	userSetCmd.Flags().StringP("telephone", "t", "", "new telephone")
}
```

如果Argument为必须项，记得使用`MarkFlagRequired`函数标记。

接下来我们实现Controller。在`controller/controller.go`文件中，定义了一个控制器类型。这是我们命令集控制器的基类型，其中`Ctx`负责上下文内容存储，其余两个成员均为Cobra中调用函数的参数类型。

```go
type Controller struct {
	Args []string
	Cmd  *cobra.Command
	Ctx  Ctx
	Srv  service.Manager
}
```

所有的命令集共用同一个`Controller`变量，以达到节省内存空间的效果。这是因为CLI程序每次运行`Run`属性所描述的函数只会调用其中一个，`Controller`也只会存储`Run`属性函数的参数，而不会存储`PersistentPreRun`、`PersistentPostRun`的参数。

唯一的`Controller`变量定义在`controller/controller.go`文件中。

```go
var ctrl Controller
```

为了区分不同命令集，需要针对每个命令集定义独有的接口。

```go
type UserCtrl interface {
	UserDelete()
	UserList()
	UserSet()
}
```

接下来，我们就可以声明Controller所要绑定的方法。

```go
func (c *Controller) UserDelete() {}

func (c *Controller) UserList() {}

func (c *Controller) UserSet() {}
```

针对每个命令集，只需通过其控制器接口来获取控制器实例。

```go
func GetUserCtrl() UserCtrl {
	return &ctrl
}
```

那么，Controller如何与Command连接起来呢？可以观察到Controller所绑定的方法均是没有参数输入的，而Command指定的函数需要传递两个参数。

在`controller/controller.go`文件中，我们定义一个封装函数，该函数用于将Controller绑定的方法和初始化一起封装成一个满足Command所需要的函数。

```go
func WrapperRun(fn func()) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		cmdStr := cmd.Name()
		cmd.VisitParents(func(pcmd *cobra.Command) { cmdStr = pcmd.Name() + "." + cmdStr })
		log.SetCommand(cmdStr)
		if len(args) != 0 {
			log.AddParams("args", args)
		}
		ctrl.Args = args
		ctrl.Cmd = cmd
		ctrl.Ctx.User = &user{
			get: func() string {
				name := ctrl.Srv.Admin().GetCurrentUserName()
				log.SetUser(name)
				return name
			},
			set: func(name string) error {
				err := ctrl.Srv.Admin().SetCurrentUserName(name)
				if err == nil && name != "" {
					log.SetUser(name)
				}
				return err
			},
		}
		ctrl.Ctx.User.Get()
		ctrl.Ctx.Value = viper.New()
		ctrl.Ctx.Value.BindPFlags(cmd.Flags())
		ctrl.Ctx.Visit = make(map[string]bool)
		cmd.Flags().Visit(func(flag *pflag.Flag) { ctrl.Ctx.Visit[flag.Name] = true })
		fn()
	}
}
```

在`cmd/wrapper.go`文件中，我们将该方法再次封装在`cmd`包中。

```go
func wrapper(fn func()) func(*cobra.Command, []string) {
	return controller.WrapperRun(fn)
}
```

在声明命令的时候，只需要封装相关的函数即可。

```go
var userListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Run:     wrapper(controller.GetUserCtrl().UserList),
}
```

### 如何使用命令行参数？

> Go Agenda使用Viper绑定和管理Arguments和Flags。

Go Agenda封装了Command所需要的函数，并在封装中加入了初始化代码。

```go
// controller/controller.go
func WrapperRun(fn func()) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// more code
	}
}
```

普通参数将被赋值到`Args`中，而带Flag的参数将被绑定到`Ctx`中。

如果有如下命令：

```sh
$ agenda user list hello -u MegaShow what happen?
```

并且相应的命令的逻辑如下：

```go
func (c *Controller) List() {
    user, _ := c.Ctx.GetString("user")
    password, _ := c.Ctx.GetSecretString("password")
	fmt.Println("user:", user)
	fmt.Println("password:", password)
	fmt.Println("others:", c.Args)
}
```

将得到如下的结果：

```
user: MegaShow
password:
others: [hello what happen?]
```

**注意：虽然初始化同时也将`Cmd`传值，但是不建议直接操作该成员。**

`Ctx`是一个被封装的`viper.Viper`类，实现了Flag封装和状态封装，根据需求决定是否记录到日志中。

```go
type Ctx struct {
	Value *viper.Viper
	Visit map[string]bool
	User  User
}
```

`Value`值的获取，分别提供了普通取值方式和`Secret`取值方式。普通方式将会将该值记录到日志中，而秘密取值不会，秘密取值适合于密码参数获取。

```go
user, userOk := c.Ctx.GetString("user")
password, passwordOk := c.Ctx.GetSecretString("password")
```

取值函数的第二个参数为其`Visit`值，类型为`bool`。

当前状态通过`c.Ctx.User`修改，这也是一个封装类型，它拥有直接与`service.Admin()`交互的权限。**因为采用了提前处理状态，服务中相关方法将被禁止调用。**

```go
currentUser := c.Ctx.User.Get()
err := c.Ctx.User.Set(user)
if err != nil {
	// more code
}
```

> 当前状态的获取和判断，应该在参数验证之后、进行服务之前。

### 如何验证参数合法性？

> 参数必选和可选利用Cobra的`MarkFlagRequired`验证，而参数合法必须自己实现逻辑验证。

本节以命令`agenda register`为例，该命令用法如下。

```sh
$ agenda help register

Usage:
  agenda register [flags]

Aliases:
  register, r, reg

Flags:
  -e, --email string       email of your new account
  -h, --help               help for register
  -p, --password string    password of your new account
  -t, --telephone string   telephone of your new account
  -u, --user string        username of your new account
```

由于部分参数在多个命令中均出现，因此我们将在`controller/verify.go`中封装参数的验证。比如，用户名的验证如下。

```go
func verifyUser(user string) {
	if user == "" {
		return
	}
	log.Verbose("check if parameter user matches rules")
	verify.AssertLength(1, 32, user, "user name too long")
	verify.AssertReg(`^[a-zA-Z][a-zA-Z0-9_]{0,31}$`, user, "user name invalid")
}
```

这里使用了封装的`verify`包，该包位于`lib/verify`文件夹中。该包的验证一旦不通过，程序将会终止。

参数的验证应该在Controller一开始就执行，必须在调用Service相关方法之前执行。验证的过程不受日志管理，即验证不通过，该记录不会存储在日志中。

```go
func (c *Controller) Register() {
	user, _ := c.Ctx.GetString("user")
	password, _ := c.Ctx.GetSecretString("password")
	email, _ := c.Ctx.GetString("email")
	telephone, _ := c.Ctx.GetString("telephone")

	verifyUser(user)
	verifyPassword(password)
	verifyEmail(email)
	verifyTelephone(telephone)
	verifyEmptyArgs(c.Args)

	err := c.Srv.Admin().Register(user, password, email, telephone)
	if err != nil {
		log.Error(err.Error())
	}
	log.Info("register account successfully")
}
```

上面的代码实现中，日志管理是在参数验证之后才开始进行的。

### 如何进行日志管理？

> Go Agenda的日志管理包对`logrus`进行了封装。

Go Agenda的日志管理在Controller验证参数之后进行。日志管理需要设置用户信息、参数，以存储日志记录。但是，密码并不在参数的范畴内。

Go Agenda的日志分为4个等级：

* `Verbose`：细节日志，记录每个执行函数操作。该日志不会被存储，且只有声明了`-v`全局Flag才会打印到标准输出中。用于调试。
* ~~`Show`：展示日志，本质上就是`fmt.Println`，记录有用信息。该日志不会被存储，只会被打印到标准输出中。~~
* `Info`：信息日志，记录命令执行成功的结果。该日志会被存储，且会被打印到标准输出中。
* `Error`：错误日志，与`Info`相反，记录命令执行失败的结果。该日志会被存储，且会被打印到标准输出中。

> ~~Go Agenda使用`Show`日志来替代`fmt.Println`，如果有更好的解决方案不妨提出来。~~
>
> Go Agenda使用`fmt.Println`替代原有的`Show`日志。

给日志设置用户名以及参数已经封装在`Ctx`和`Wrapper`中，只要按照规定获取参数，即可正常工作。

### 如何编写一个服务？

> Go Agenda的服务均绑定在同一个Service类型上，但使用管理者Manager来隔离绑定的方法。

在`service/service.go`中，定义了类型Service和Manager。

```go
type Service struct {
	DB model.Manager
}

type Manager Service

func (s *Manager) GetService() *Service {
	return (*Service)(s)
}
```

编写服务时，将服务函数均绑定在Service上，但是通过Manager的相应的方法暴露出去。

Controller中，拥有成员`service.Manager`。

```go
type Controller struct {
	Args    []string
	Cmd     *cobra.Command
	Ctx     *viper.Viper
	Srv		service.Manager
}
```

在每一个服务`.go`文件中，均定义了获取相应服务集的方法。

```go
func (s *Manager) Admin() AdminService {
	return s.GetService()
}
```

这里的`AdminService`是一个接口，其中定义了相应服务的方法。

```go
type AdminService interface {
	Login(name, password string) error
	Register(name, password, email, telephone string) error
}
```

### 如何编写一个数据库模型？

与服务相似，数据库模型也采用了本体加管理者的方式实现。

注意，数据库使用前必须手动初始化，使用后必须手动回收。

不过，由于采用了管理者来进行数据库的管理，可以将初始化交给管理者负责。

```go
func (m *Manager) User() UserModel {
	if userDB.isInit == false {
		userDB.initModel(&userDB.Data)
	}
	return &userDB
}
```

回收释放的代码位于`controller/controller.go`中，由`rootCmd`负责调用。

```go
func CtrlRelease(cmd *cobra.Command, args []string) {
	log.Release()
	model.ReleaseUserModel()
	model.ReleaseMeetingModel()
	model.ReleaseStatusModel()
}
```

### 其他

请询问，可继续添加。

