from datetime import datetime, timedelta

routines = ['Hamstring', 'Hip', 'Posture', 'Shoulder']
weeks = range(1, 5)  # 1-4 weeks
days = range(1, 4)   # 1-3 days per week

start_date = datetime.now()
current_date = start_date
days_count = 0
routine_index = 0
week = 1
day = 1

output = []

while days_count < 154:
    routine = routines[routine_index]
    
    # Format the output
    output.append(f"- description: {routine} W{week}D{day}")
    output.append(f"  date: {current_date.strftime('%Y-%m-%d')}")
    output.append("")
    
    # Move to next routine
    routine_index = (routine_index + 1) % len(routines)
    
    # If we've gone through all routines, move to next day
    if routine_index == 0:
        day += 1
        if day > 3:
            day = 1
            week += 1
            if week > 4:
                week = 1
    
    current_date += timedelta(days=1)
    days_count += 1

# Print the output
print("\n".join(output))
