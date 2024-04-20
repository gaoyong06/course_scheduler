#### TODO

- [x] 验证初始化种群质量
- [ ] 验证交叉逻辑
- [ ] 验证变异逻辑
- [ ] 验证适应度收敛

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
4. 