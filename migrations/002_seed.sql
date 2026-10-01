-- Seed data for the dev database. IDs are predictable because the tables are fresh.

INSERT INTO authors (name, bio) VALUES
    ('Ursula K. Le Guin', 'Earthsea, Hainish Cycle.'),        -- 1
    ('Octavia E. Butler', 'Kindred, Patternist series.'),     -- 2
    ('Terry Pratchett',   'Discworld.'),                      -- 3
    ('Ted Chiang',        NULL),                              -- 4
    ('N. K. Jemisin',     'The Broken Earth trilogy.'),       -- 5
    ('Anonymous',         'Has no books. Useful for LEFT JOIN practice.'); -- 6

INSERT INTO books (author_id, title, isbn, published_year, total_copies, available_copies) VALUES
    (1, 'A Wizard of Earthsea',            '9780553383041', 1968, 3, 2),  -- 1
    (1, 'The Left Hand of Darkness',       '9780441478125', 1969, 2, 2),  -- 2
    (1, 'The Dispossessed',                '9780061054884', 1974, 1, 0),  -- 3
    (2, 'Kindred',                         '9780807083697', 1979, 4, 3),  -- 4
    (2, 'Parable of the Sower',            '9780446675505', 1993, 2, 1),  -- 5
    (2, 'Dawn',                            '9780446603775', 1987, 1, 1),  -- 6
    (3, 'Guards! Guards!',                 '9780062225757', 1989, 2, 1),  -- 7
    (3, 'Small Gods',                      '9780062237378', 1992, 1, 1),  -- 8
    (3, 'Going Postal',                    '9780060502935', 2004, 2, 2),  -- 9
    (4, 'Stories of Your Life and Others', '9781101972120', 2002, 2, 1),  -- 10
    (4, 'Exhalation',                      '9781101947883', 2019, 3, 3),  -- 11
    (5, 'The Fifth Season',                '9780316229296', 2015, 2, 0),  -- 12
    (5, 'The Broken Earth: Companion',     'TBD',           NULL, 1, 1),  -- 13
    (3, 'Small Gods (Reprint)',            '9780062237378', 2013, 1, 1);  -- 14

INSERT INTO members (email, full_name) VALUES
    ('ada@example.com',    'Ada Lovelace'),     -- 1
    ('grace@example.com',  'Grace Hopper'),     -- 2
    ('alan@example.com',   'Alan Turing'),      -- 3
    ('edsger@example.com', 'Edsger Dijkstra'),  -- 4
    ('barbara@example.com','Barbara Liskov');   -- 5 (never borrowed anything)

INSERT INTO loans (book_id, member_id, borrowed_at, due_at, returned_at) VALUES
    -- active loans
    (1,  1, now() - interval '5 days',   now() + interval '9 days',  NULL),
    (3,  1, now() - interval '30 days',  now() - interval '16 days', NULL),  -- overdue
    (12, 1, now() - interval '2 days',   now() + interval '12 days', NULL),
    (4,  2, now() - interval '10 days',  now() + interval '4 days',  NULL),
    (5,  2, now() - interval '20 days',  now() - interval '6 days',  NULL),  -- overdue
    (7,  3, now() - interval '1 day',    now() + interval '13 days', NULL),
    (10, 3, now() - interval '40 days',  now() - interval '26 days', NULL),  -- overdue
    (12, 4, now() - interval '3 days',   now() + interval '11 days', NULL),
    -- returned loans
    (2,  2, now() - interval '60 days',  now() - interval '46 days', now() - interval '50 days'),
    (4,  3, now() - interval '90 days',  now() - interval '76 days', now() - interval '70 days'),
    (9,  4, now() - interval '45 days',  now() - interval '31 days', now() - interval '25 days'), -- returned late
    (11, 1, now() - interval '100 days', now() - interval '86 days', now() - interval '95 days'),
    (1,  4, now() - interval '30 days',  now() - interval '16 days', now() - interval '18 days');
