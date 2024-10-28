import tkinter as tk
import random
from PIL import Image, ImageTk


class WhacAMoleGame:
    def __init__(self, root):
        self.root = root
        self.root.title("打地鼠游戏")
        self.score = 0
        self.buttons = []
        self.mole_positions = {}  # 存储当前出现地鼠的按钮索引和类型（地鼠或炸弹）
        self.time_limit = 30000  # 游戏总时长为 30 秒
        self.mole_duration = 1000  # 地鼠显示时长为 1 秒
        self.mole_count = 3  # 同时出现的地鼠数量

        # 设置窗口大小
        x,y=self.root.maxsize()
        self.root.geometry(f"{int(x*0.5)}x{int(y*0.5)}")
        self.root.resizable(False, False)#不可缩放屏幕大小

        # 加载地鼠和炸弹图片
        self.mole_image = Image.open("486.jpg").resize((80, 80))
        self.mole_photo = ImageTk.PhotoImage(self.mole_image)
        self.bomb_image = Image.open("Long.jpg").resize((80, 80))
        self.bomb_photo = ImageTk.PhotoImage(self.bomb_image)

        self.create_widgets()
        self.start_game()

    def create_widgets(self):
        button_width = 1 / 3
        button_height = 1 / 3

        for i in range(9):
            button = tk.Button(self.root,image="",command=self.create_button_command(i))
            row = i // 3
            col = i % 3
            button.place(relx=col * button_width,rely=row * button_height,relwidth=button_width,relheight=button_height)
            self.buttons.append(button)

        # 创建得分标签
        self.score_label = tk.Label(self.root, text=f"分数: {self.score}", font=("Arial", 16))
        self.score_label.place(relx=0.5, rely=0.9, anchor="center")

    def create_button_command(self, index):
        def button_command():
            self.hit_mole(index)

        return button_command

    def spawn_moles(self):
        # 移除当前地鼠
        for index in list(self.mole_positions.keys()):
            self.buttons[index].config(image="")
            del self.mole_positions[index]

        # 随机选择几个按钮显示地鼠或炸弹
        Realcount=random.randint(1,self.mole_count)
        for _ in range(Realcount):
            index = random.randint(0, 8)
            is_bomb = random.choice([True, False])  # 随机选择是地鼠还是炸弹
            if is_bomb:
                self.buttons[index].config(image=self.bomb_photo)
                self.mole_positions[index] = "bomb"
            else:
                self.buttons[index].config(image=self.mole_photo)
                self.mole_positions[index] = "mole"

        # 1 秒后移除这些地鼠和炸弹
        self.root.after(self.mole_duration, self.clear_moles)

    def clear_moles(self):
        # 清空所有当前地鼠或炸弹
        for index in self.mole_positions.keys():
            self.buttons[index].config(image="")
        self.mole_positions.clear()
        self.spawn_moles()  # 生成新的地鼠

    def hit_mole(self, index):
        if index in self.mole_positions:
            if self.mole_positions[index] == "mole":
                self.score += 1
            elif self.mole_positions[index] == "bomb":
                self.score -= 10
            # 更新分数并清除该按钮
            self.score_label.config(text=f"分数: {self.score}")
            self.buttons[index].config(image="")
            del self.mole_positions[index]
        else:
            # 点击空白处扣分
            self.score -= 1
            self.score_label.config(text=f"分数: {self.score}")

    def start_game(self):
        self.spawn_moles()
        self.root.after(self.time_limit, self.end_game)

    def end_game(self):
        for button in self.buttons:
            button.config(state="disabled")
        for index in self.mole_positions:
            self.buttons[index].config(image="")
        self.score_label.config(text=f"游戏结束！最终分数: {self.score}")


# 运行游戏
root = tk.Tk()
game = WhacAMoleGame(root)
root.mainloop()
