import yaml
from datetime import datetime, timedelta

def get_next_date(current_date, target_days):
    while current_date.weekday() not in target_days:
        current_date += timedelta(days=1)
    return current_date

def generate_core_sessions():
    # Start from tomorrow
    current_date = datetime.now().date() + timedelta(days=1)
    end_date = datetime(2025, 5, 31).date()
    
    # Wednesday is 2, Sunday is 6
    target_days = {2, 6}  
    
    sessions = []
    current_date = get_next_date(current_date, target_days)
    
    while current_date <= end_date:
        sessions.append({
            'description': 'Stone Circle',
            'date': datetime.combine(current_date, datetime.min.time()).strftime('%Y-%m-%dT00:00:00Z')
        })
        
        # Move to next day and find next target day
        current_date += timedelta(days=1)
        current_date = get_next_date(current_date, target_days)
    
    return {'sessions': sessions}

if __name__ == '__main__':
    plan = generate_core_sessions()
    print(yaml.dump(plan, sort_keys=False))
