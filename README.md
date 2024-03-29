### 政府网站采集器

给定一个政府网站网址 可选择一些配置让采集更准确
1. 先采网站地图可以先寻找网站robots.txt
2. 在根据网站地图进行筛选
3. 根据地图获取文章
4. 根据采到的文章分类 计算数量
5. 根据分好类的类型采集对应文章内容等需求字段

#### 4个配置表格
1. gov_map_settings 配置采集map规则
2. gov_art_settings 配置采集文章规则(只采集了url链接! 主要配置翻页text和html是否是html渲染的)
3. gov_art_rules 配置文章字段xpath规则
4. gov_art_update_settings 配置文章字段更新规则

### 测试站点:
 - [http://www.qiaokou.gov.cn/](http://www.qiaokou.gov.cn/)
 - [http://www.jianghan.gov.cn/](http://www.jianghan.gov.cn/)
 - [http://www.jiangan.gov.cn/](http://www.jiangan.gov.cn/)
 - [http://www.hanyang.gov.cn/](http://www.hanyang.gov.cn/)
 - [http://www.wuchang.gov.cn/](http://www.wuchang.gov.cn/)
 - [http://www.qingshan.gov.cn/](http://www.qingshan.gov.cn/)
 - [http://www.hongshan.gov.cn/](http://www.hongshan.gov.cn/)
 - [http://www.dxh.gov.cn/](http://www.dxh.gov.cn/)
 - [http://www.huangpi.gov.cn/](http://www.huangpi.gov.cn/)
 - [http://www.jiangxia.gov.cn/](http://www.jiangxia.gov.cn/)
 - [http://www.caidian.gov.cn/](http://www.caidian.gov.cn/)
 - [http://www.wedz.com.cn/](http://www.wedz.com.cn/)
 - [http://www.whxinzhou.gov.cn/](http://www.whxinzhou.gov.cn/)
 - [http://www.hbdaye.gov.cn/](http://www.hbdaye.gov.cn/)
 - [http://www.huangshigang.gov.cn/](http://www.huangshigang.gov.cn/)
 - [http://www.xisaishan.gov.cn/](http://www.xisaishan.gov.cn/)
 - [http://www.xialuqu.gov.cn/](http://www.xialuqu.gov.cn/)
 - [http://www.hsts.gov.cn/](http://www.hsts.gov.cn/)
 - [http://www.yx.gov.cn/](http://www.yx.gov.cn/)
 - [http://www.yunxi.gov.cn/](http://www.yunxi.gov.cn/)
 - [http://www.zhushan.gov.cn/](http://www.zhushan.gov.cn/)
 - [http://www.zhuxi.gov.cn/](http://www.zhuxi.gov.cn/)
 - [http://www.fangxian.gov.cn/](http://www.fangxian.gov.cn/)
 - [http://maojian.shiyan.gov.cn/](http://maojian.shiyan.gov.cn/)
 - [http://www.zhangwan.gov.cn/](http://www.zhangwan.gov.cn/)
 - [http://www.hbyx.gov.cn/](http://www.hbyx.gov.cn/)
 - [http://www.djk.gov.cn/](http://www.djk.gov.cn/)
 - [http://www.yuanan.gov.cn/](http://www.yuanan.gov.cn/)
 - [http://www.hbxsx.gov.cn/](http://www.hbxsx.gov.cn/)
 - [http://www.hbzg.gov.cn/](http://www.hbzg.gov.cn/)
 - [http://www.ycxl.gov.cn/](http://www.ycxl.gov.cn/)
 - [http://www.ycwjg.gov.cn/](http://www.ycwjg.gov.cn/)
 - [http://www.dianjun.gov.cn/](http://www.dianjun.gov.cn/)
 - [http://www.xiaoting.gov.cn/](http://www.xiaoting.gov.cn/)
 - [http://www.changyang.gov.cn/](http://www.changyang.gov.cn/)
 - [http://www.hbwf.gov.cn/](http://www.hbwf.gov.cn/)
 - [http://www.yidu.gov.cn](http://www.yidu.gov.cn)
 - [http://www.dangyang.gov.cn/](http://www.dangyang.gov.cn/)
 - [http://www.zgzhijiang.gov.cn/](http://www.zgzhijiang.gov.cn/)
 - [http://www.10.gov.cn/](http://www.10.gov.cn/)
 - [http://www.yichang.gov.cn/](http://www.yichang.gov.cn/)
 - [http://www.zyzf.gov.cn/](http://www.zyzf.gov.cn/)
 - [http://www.ych.gov.cn/](http://www.ych.gov.cn/)
 - [http://www.lhk.gov.cn/](http://www.lhk.gov.cn/)
 - [http://www.xfxc.gov.cn/](http://www.xfxc.gov.cn/)
 - [http://www.fc.gov.cn/](http://www.fc.gov.cn/)
 - [http://www.xyxz.gov.cn/](http://www.xyxz.gov.cn/)
 - [http://www.hbnz.gov.cn/](http://www.hbnz.gov.cn/)
 - [http://www.baokang.gov.cn/](http://www.baokang.gov.cn/)
 - [http://www.hbgucheng.gov.cn/](http://www.hbgucheng.gov.cn/)
 - [http://www.yingcheng.gov.cn/](http://www.yingcheng.gov.cn/)
 - [http://www.hbdawu.gov.cn/](http://www.hbdawu.gov.cn/)
 - [http://www.xiaochang.gov.cn/](http://www.xiaochang.gov.cn/)
 - [http://www.yunmeng.gov.cn/](http://www.yunmeng.gov.cn/)
 - [http://www.anlu.gov.cn/](http://www.anlu.gov.cn/)
 - [http://www.hanchuan.gov.cn/](http://www.hanchuan.gov.cn/)
 - [http://www.xiaonan.gov.cn/](http://www.xiaonan.gov.cn/)
 - [http://www.jiangling.gov.cn/](http://www.jiangling.gov.cn/)
 - [http://www.gongan.gov.cn/](http://www.gongan.gov.cn/)
 - [http://www.hbsz.gov.cn/](http://www.hbsz.gov.cn/)
 - [http://www.shishou.gov.cn/](http://www.shishou.gov.cn/)
 - [http://www.jianli.gov.cn/](http://www.jianli.gov.cn/)
 - [http://www.shashi.gov.cn/](http://www.shashi.gov.cn/)
 - [http://www.honghu.gov.cn/](http://www.honghu.gov.cn/)
 - [http://www.huangzhou.gov.cn/](http://www.huangzhou.gov.cn/)
 - [http://www.wuxue.gov.cn/](http://www.wuxue.gov.cn/)
 - [http://www.macheng.gov.cn/](http://www.macheng.gov.cn/)
 - [http://www.hazf.gov.cn/](http://www.hazf.gov.cn/)
 - [http://www.luotian.gov.cn/](http://www.luotian.gov.cn/)
 - [http://www.chinays.gov.cn/](http://www.chinays.gov.cn/)
 - [http://www.xishui.gov.cn/](http://www.xishui.gov.cn/)
 - [http://www.qichun.gov.cn/](http://www.qichun.gov.cn/)
 - [http://www.hm.gov.cn/](http://www.hm.gov.cn/)
 - [http://www.tfzf.gov.cn/](http://www.tfzf.gov.cn/)
 - [http://www.liangzh.gov.cn/](http://www.liangzh.gov.cn/)
 - [http://www.hbhr.gov.cn/](http://www.hbhr.gov.cn/)
 - [http://www.echeng.gov.cn/](http://www.echeng.gov.cn/)
 - [http://www.jingshan.gov.cn/](http://www.jingshan.gov.cn/)
 - [http://www.shayang.gov.cn/](http://www.shayang.gov.cn/)
 - [http://www.jmdbq.gov.cn/](http://www.jmdbq.gov.cn/)
 - [http://www.duodao.gov.cn/](http://www.duodao.gov.cn/)
 - [http://www.zhongxiang.gov.cn/](http://www.zhongxiang.gov.cn/)
 - [http://www.jiayu.gov.cn/](http://www.jiayu.gov.cn/)
 - [http://www.zgtc.gov.cn/](http://www.zgtc.gov.cn/)
 - [http://www.chongyang.gov.cn/](http://www.chongyang.gov.cn/)
 - [http://www.tongshan.gov.cn/](http://www.tongshan.gov.cn/)
 - [http://www.xianan.gov.cn/](http://www.xianan.gov.cn/)
 - [http://www.chibi.gov.cn/](http://www.chibi.gov.cn/)
 - [http://www.zggsw.gov.cn/](http://www.zggsw.gov.cn/)
 - [http://www.zgsuixian.gov.cn/](http://www.zgsuixian.gov.cn/)
 - [http://www.zengdu.gov.cn/](http://www.zengdu.gov.cn/)
 - [https://es.gov.cn/](https://es.gov.cn/)
 - [http://www.lichuan.gov.cn/](http://www.lichuan.gov.cn/)
 - [http://www.hbjs.gov.cn/](http://www.hbjs.gov.cn/)
 - [http://www.hbbd.gov.cn/](http://www.hbbd.gov.cn/)
 - [http://www.xianfeng.gov.cn/](http://www.xianfeng.gov.cn/)
 - [http://www.xe.gov.cn/](http://www.xe.gov.cn/)
 - [http://www.xe.gov.cn/](http://www.xe.gov.cn/)
 - [http://www.hefeng.gov.cn/](http://www.hefeng.gov.cn/)
 - [http://www.xiantao.gov.cn/](http://www.xiantao.gov.cn/)
 - [http://www.hbqj.gov.cn/](http://www.hbqj.gov.cn/)
 - [http://www.tianmen.gov.cn/](http://www.tianmen.gov.cn/)
 - [http://www.snj.gov.cn/](http://www.snj.gov.cn/)

