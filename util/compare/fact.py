def хэвлэ(тоо):
    print(тоо)

def факториал(н):
    if н <= 1:
        return 1

    бодсон = н * факториал(н - 1)
    хэвлэ(бодсон)
    return бодсон

def үндсэн():
    return факториал(20)

# Run the program
if __name__ == "__main__":
    result = үндсэн()
    print("Final result:", result)
