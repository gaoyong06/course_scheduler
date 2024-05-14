#### TODO

- [x] 验证初始化种群质量
- [x] 验证交叉逻辑
- [x] 验证变异逻辑
- [x] 验证适应度收敛
- [x] 遗传算法退出条件定义
- [x] 输入条件定义
- [] 输出结果定义

#### 业务流程

1. 根据年级,班级,科目 生成课班信息Classes(SN, 科目, 年级, 班级, 名称)
2. 获取各个科目的周课时
3. 初始化种群population
   1. 打乱课班Classes顺序
   2. 初始化课班适应性矩阵classMatrix([课班(科目_年级_班级)][教师][教室][时间段], value: Val)
   3. 计算课班适应性矩阵classMatrix匹配结果值(即: 为classMatrix的value Val的score赋值)
   4. 课班适应性矩阵classMatrix分配(即: 为classMatrix的value Val的Used赋值)
   5. 根据课班适应性矩阵classMatrix分配结果生成个体Individual
   6. 将生成的个体Individual添加至种群population
   7. 回到3.1, 继续执行,连续执行populationSize次, 使种群中有populationSize个个体


#### 约束条件

约束条件分为: 固定约束条件, 动态约束条件

区分固定约束条件和动态约束条件的最重要标准是：固定约束条件是在排课之前就已知的，不需要根据当前已排的课程情况来判断是否满足约束；而动态约束条件则需要在排课过程中动态地检查和更新,在排课过程中实时检查已排的课程和待排的课程之间的关系，以确保排课结果满足这个约束条件。

在排课程序中，需要区分固定约束条件和动态约束条件，因为这两种约束条件在业务处理方式上有所不同。固定约束条件通常在排课之前就已知，可以提前进行检查和处理，而动态约束条件则需要在排课过程中不断地检查和更新。因此，在设计排课算法时，需要根据具体的约束条件，采用不同的处理方式，以提高算法的效率和精度。

例如，对于固定约束条件，可以在排课之前就对其进行检查和处理，将不满足约束条件的课程排除在外，或者对其进行特殊处理，以减少后续排课过程中的检查次数。而对于动态约束条件，则需要在排课过程中不断地检查已排的课程和待排的课程之间的关系，以确保排课结果满足这个约束条件。这可能需要不断地更新约束条件的状态，以反映已排的课程对待排的课程的影响。因此，对于动态约束条件，排课算法需要更加灵活和高效，以适应不断变化的约束条件。

固定约束条件通常是事先已知的、不会随时间变化的规则，例如：

1. 每门课程每周的上课时长
2. 每个教师每周可上课的时长
3. 每个教室每周可使用的时长
4. 某些时间段不能安排某些课程或教师

动态约束条件则是随着排课过程的进行而变化的规则，例如：

1. 同一时间段不能安排重复的课程或教师
2. 某些课程需要先后次序进行安排
3. 某些课程需要间隔一定时间才能进行安排

对于固定约束条件，您可以在初始化阶段就对其进行处理，例如给不满足条件的时间段赋予较低的分数，或者直接将其标记为不可用。对于动态约束条件，您可以在每次迭代的排课过程中检查当前已安排的课程，并根据当前状态动态调整分数或可用性。


#### 规划
1. 用户交互和反馈：允许用户（例如学生和教师）在排课过程中提供反馈，并根据反馈进行调整。这可以通过为用户提供一个图形化的界面来实现，用户可以在界面上查看当前的排课方案，并提供有关偏好和需求的反馈。
2. 数据分析和可视化：使用数据分析和可视化工具来分析排课过程中产生的数据，以帮助识别潜在问题和改进排课策略。例如，您可以使用数据可视化工具来显示课程的时间和空间分布，以便更好地理解排课情况并识别可能的问题。


#### 算法程序提供以下几个路由
1. 接收数据：接收网站程序发送过来的数据，包括课程信息、教师信息、教室信息等
2. 执行排课：根据接收到的数据执行排课算法，并返回排课结果
3. 查询排课结果：根据查询条件查询排课结果，例如根据课程名称查询排课结果

##### 接收数据
1. 接收网站程序POST提交过来的数据,包括课程信息、教师信息、教室信息等
2. 在排课任务队列中新增一条排课任务,task_data设置为网站程序提交的数据, status为pending
3. 新增一条排课任务后, 给网站程序返回一个task_id 作为接收数据的返回值

##### 执行排课
3. 处理任务队列程序从任务队列中获取到该任务,根据task_data内部的数据,执行排课
4. 排课完成后,将排课任务的status修改为success或者failed
5. 排课完成后,将排课结果写入排课结果数据表

##### 查询排课结果
1. 网站程序根据排课任务ID,来排课程序处查询排课结果,如果任务状态是success,则返回排课结果,如果是pending,running,failed则排课结果为空